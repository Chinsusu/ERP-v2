package auth

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	defaultAccessTokenTTL  = 8 * time.Hour
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
	defaultLockoutWindow   = 15 * time.Minute
	defaultLockoutDuration = 15 * time.Minute
	defaultMinPasswordLen  = 10
	defaultMaxFailedLogins = 5
)

type PasswordPolicy struct {
	MinLength              int
	RequireLetter          bool
	RequireNumberOrSymbol  bool
	CommonPasswordsBlocked bool
}

type LockoutPolicy struct {
	MaxFailedAttempts int
	Window            time.Duration
	Duration          time.Duration
}

type LoginFailureCode string

const (
	LoginFailureInvalidCredentials LoginFailureCode = "invalid_credentials"
	LoginFailurePasswordPolicy     LoginFailureCode = "password_policy"
	LoginFailureLocked             LoginFailureCode = "locked"
)

type LoginFailure struct {
	Code        LoginFailureCode
	Message     string
	LockedUntil time.Time
}

type Session struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	Principal        Principal
}

type failedLoginState struct {
	Attempts    int
	FirstFailed time.Time
	LockedUntil time.Time
}

type SessionManager struct {
	cfg           MockConfig
	password      PasswordPolicy
	lockout       LockoutPolicy
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
	mu            sync.Mutex
	accessTokens  map[string]Session
	refreshTokens map[string]Session
	failedLogins  map[string]failedLoginState
}

func NewSessionManager(cfg MockConfig, now func() time.Time) *SessionManager {
	if now == nil {
		now = time.Now
	}

	manager := &SessionManager{
		cfg: cfg,
		password: PasswordPolicy{
			MinLength:              defaultMinPasswordLen,
			RequireLetter:          true,
			RequireNumberOrSymbol:  true,
			CommonPasswordsBlocked: true,
		},
		lockout: LockoutPolicy{
			MaxFailedAttempts: defaultMaxFailedLogins,
			Window:            defaultLockoutWindow,
			Duration:          defaultLockoutDuration,
		},
		accessTTL:     defaultAccessTokenTTL,
		refreshTTL:    defaultRefreshTokenTTL,
		now:           now,
		accessTokens:  make(map[string]Session),
		refreshTokens: make(map[string]Session),
		failedLogins:  make(map[string]failedLoginState),
	}

	manager.seedStaticAccessToken()
	return manager
}

func (m *SessionManager) PasswordPolicy() PasswordPolicy {
	return m.password
}

func (m *SessionManager) LockoutPolicy() LockoutPolicy {
	return m.lockout
}

func (m *SessionManager) Login(email string, password string) (Session, LoginFailure, bool) {
	normalizedEmail := normalizeEmail(email)
	now := m.now().UTC()

	m.mu.Lock()
	if lockedUntil, locked := m.lockedUntil(normalizedEmail, now); locked {
		m.mu.Unlock()
		return Session{}, LoginFailure{
			Code:        LoginFailureLocked,
			Message:     "Account temporarily locked after repeated failed login attempts",
			LockedUntil: lockedUntil,
		}, false
	}
	m.mu.Unlock()

	if failure := ValidatePasswordPolicy(password, m.password); failure != "" {
		m.recordFailedLogin(normalizedEmail, now)
		return Session{}, LoginFailure{
			Code:    LoginFailurePasswordPolicy,
			Message: failure,
		}, false
	}

	principal, ok := ValidateMockLogin(m.cfg, email, password)
	if !ok {
		lockedUntil := m.recordFailedLogin(normalizedEmail, now)
		failure := LoginFailure{
			Code:    LoginFailureInvalidCredentials,
			Message: "Invalid email or password",
		}
		if !lockedUntil.IsZero() {
			failure.Code = LoginFailureLocked
			failure.Message = "Account temporarily locked after repeated failed login attempts"
			failure.LockedUntil = lockedUntil
		}
		return Session{}, failure, false
	}

	session := m.issueSession(principal, now)
	m.clearFailedLogin(normalizedEmail)
	return session, LoginFailure{}, true
}

func (m *SessionManager) Refresh(refreshToken string) (Session, bool) {
	now := m.now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.refreshTokens[strings.TrimSpace(refreshToken)]
	if !ok || !existing.RefreshExpiresAt.After(now) {
		return Session{}, false
	}

	delete(m.refreshTokens, existing.RefreshToken)
	delete(m.accessTokens, existing.AccessToken)

	session := m.newSessionLocked(existing.Principal, now)
	m.storeSessionLocked(session)
	return session, true
}

func (m *SessionManager) AuthenticateAccessToken(accessToken string) (Principal, bool) {
	now := m.now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.accessTokens[strings.TrimSpace(accessToken)]
	if !ok || !session.AccessExpiresAt.After(now) {
		return Principal{}, false
	}

	return session.Principal, true
}

func (m *SessionManager) issueSession(principal Principal, now time.Time) Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := m.newSessionLocked(principal, now)
	m.storeSessionLocked(session)
	return session
}

func (m *SessionManager) newSessionLocked(principal Principal, now time.Time) Session {
	return Session{
		AccessToken:      "local-at-" + randomToken(),
		RefreshToken:     "local-rt-" + randomToken(),
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshExpiresAt: now.Add(m.refreshTTL),
		Principal:        principal,
	}
}

func (m *SessionManager) storeSessionLocked(session Session) {
	m.accessTokens[session.AccessToken] = session
	m.refreshTokens[session.RefreshToken] = session
}

func (m *SessionManager) seedStaticAccessToken() {
	if strings.TrimSpace(m.cfg.AccessToken) == "" {
		return
	}

	now := m.now().UTC()
	session := Session{
		AccessToken:      m.cfg.AccessToken,
		RefreshToken:     "local-rt-static",
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshExpiresAt: now.Add(m.refreshTTL),
		Principal:        MockPrincipal(m.cfg),
	}

	m.accessTokens[session.AccessToken] = session
	m.refreshTokens[session.RefreshToken] = session
}

func (m *SessionManager) lockedUntil(email string, now time.Time) (time.Time, bool) {
	state, ok := m.failedLogins[email]
	if !ok || state.LockedUntil.IsZero() {
		return time.Time{}, false
	}
	if state.LockedUntil.After(now) {
		return state.LockedUntil, true
	}

	delete(m.failedLogins, email)
	return time.Time{}, false
}

func (m *SessionManager) recordFailedLogin(email string, now time.Time) time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.failedLogins[email]
	if state.FirstFailed.IsZero() || now.Sub(state.FirstFailed) > m.lockout.Window {
		state = failedLoginState{FirstFailed: now}
	}

	state.Attempts++
	if state.Attempts >= m.lockout.MaxFailedAttempts {
		state.LockedUntil = now.Add(m.lockout.Duration)
	}

	m.failedLogins[email] = state
	return state.LockedUntil
}

func (m *SessionManager) clearFailedLogin(email string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.failedLogins, email)
}

func ValidatePasswordPolicy(password string, policy PasswordPolicy) string {
	if len(password) < policy.MinLength {
		return "Password does not meet the minimum length policy"
	}
	if policy.CommonPasswordsBlocked && isCommonPassword(password) {
		return "Password is too common"
	}

	var hasLetter bool
	var hasNumberOrSymbol bool
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) || unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasNumberOrSymbol = true
		}
	}

	if policy.RequireLetter && !hasLetter {
		return "Password must include at least one letter"
	}
	if policy.RequireNumberOrSymbol && !hasNumberOrSymbol {
		return "Password must include at least one number or symbol"
	}

	return ""
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isCommonPassword(password string) bool {
	switch strings.ToLower(strings.TrimSpace(password)) {
	case "password", "password1", "password123", "1234567890", "admin123456", "qwerty12345":
		return true
	default:
		return false
	}
}

func randomToken() string {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic("auth token random source unavailable")
	}
	return hex.EncodeToString(raw[:])
}
