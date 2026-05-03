package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

type SessionStore interface {
	StoreSession(session Session, now time.Time) error
	FindByAccessToken(accessToken string, now time.Time) (Session, bool, error)
	RotateRefreshToken(refreshToken string, now time.Time, buildNext func(Session) Session) (Session, bool, error)
	RevokeRefreshToken(refreshToken string, now time.Time) (bool, error)
}

type InMemorySessionStore struct {
	mu            sync.Mutex
	accessTokens  map[string]Session
	refreshTokens map[string]Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		accessTokens:  make(map[string]Session),
		refreshTokens: make(map[string]Session),
	}
}

func (s *InMemorySessionStore) StoreSession(session Session, _ time.Time) error {
	if s == nil {
		return errors.New("auth session store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.accessTokens[session.AccessToken] = session
	s.refreshTokens[session.RefreshToken] = session
	return nil
}

func (s *InMemorySessionStore) FindByAccessToken(accessToken string, now time.Time) (Session, bool, error) {
	if s == nil {
		return Session{}, false, errors.New("auth session store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.accessTokens[strings.TrimSpace(accessToken)]
	if !ok || !session.AccessExpiresAt.After(now) {
		return Session{}, false, nil
	}

	return session, true, nil
}

func (s *InMemorySessionStore) RotateRefreshToken(
	refreshToken string,
	now time.Time,
	buildNext func(Session) Session,
) (Session, bool, error) {
	if s == nil {
		return Session{}, false, errors.New("auth session store is required")
	}
	if buildNext == nil {
		return Session{}, false, errors.New("auth session rotation builder is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.refreshTokens[strings.TrimSpace(refreshToken)]
	if !ok || !existing.RefreshExpiresAt.After(now) {
		return Session{}, false, nil
	}

	session := buildNext(existing)
	delete(s.refreshTokens, existing.RefreshToken)
	delete(s.accessTokens, existing.AccessToken)
	s.accessTokens[session.AccessToken] = session
	s.refreshTokens[session.RefreshToken] = session
	return session, true, nil
}

func (s *InMemorySessionStore) RevokeRefreshToken(refreshToken string, now time.Time) (bool, error) {
	if s == nil {
		return false, errors.New("auth session store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.refreshTokens[strings.TrimSpace(refreshToken)]
	if !ok || !existing.RefreshExpiresAt.After(now) {
		return false, nil
	}

	delete(s.refreshTokens, existing.RefreshToken)
	delete(s.accessTokens, existing.AccessToken)
	return true, nil
}

type LoginFailureStore interface {
	LockedUntil(email string, now time.Time) (time.Time, bool, error)
	RecordFailure(email string, now time.Time, policy LockoutPolicy) (time.Time, error)
	Clear(email string) error
}

type InMemoryLoginFailureStore struct {
	mu           sync.Mutex
	failedLogins map[string]failedLoginState
}

func NewInMemoryLoginFailureStore() *InMemoryLoginFailureStore {
	return &InMemoryLoginFailureStore{
		failedLogins: make(map[string]failedLoginState),
	}
}

func (s *InMemoryLoginFailureStore) LockedUntil(email string, now time.Time) (time.Time, bool, error) {
	if s == nil {
		return time.Time{}, false, errors.New("auth login failure store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.failedLogins[email]
	if !ok || state.LockedUntil.IsZero() {
		return time.Time{}, false, nil
	}
	if state.LockedUntil.After(now) {
		return state.LockedUntil, true, nil
	}

	delete(s.failedLogins, email)
	return time.Time{}, false, nil
}

func (s *InMemoryLoginFailureStore) RecordFailure(
	email string,
	now time.Time,
	policy LockoutPolicy,
) (time.Time, error) {
	if s == nil {
		return time.Time{}, errors.New("auth login failure store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.failedLogins[email]
	if state.FirstFailed.IsZero() || now.Sub(state.FirstFailed) > policy.Window {
		state = failedLoginState{FirstFailed: now}
	}

	state.Attempts++
	if state.Attempts >= policy.MaxFailedAttempts {
		state.LockedUntil = now.Add(policy.Duration)
	}

	s.failedLogins[email] = state
	return state.LockedUntil, nil
}

func (s *InMemoryLoginFailureStore) Clear(email string) error {
	if s == nil {
		return errors.New("auth login failure store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.failedLogins, email)
	return nil
}

type SessionManager struct {
	cfg               MockConfig
	password          PasswordPolicy
	lockout           LockoutPolicy
	accessTTL         time.Duration
	refreshTTL        time.Duration
	now               func() time.Time
	sessionStore      SessionStore
	loginFailureStore LoginFailureStore
}

func NewSessionManager(cfg MockConfig, now func() time.Time) *SessionManager {
	manager, err := NewSessionManagerWithStores(
		cfg,
		now,
		NewInMemorySessionStore(),
		NewInMemoryLoginFailureStore(),
	)
	if err != nil {
		panic(err)
	}

	return manager
}

func NewSessionManagerWithSessionStore(cfg MockConfig, now func() time.Time, sessionStore SessionStore) (*SessionManager, error) {
	return NewSessionManagerWithStores(cfg, now, sessionStore, NewInMemoryLoginFailureStore())
}

func NewSessionManagerWithStores(
	cfg MockConfig,
	now func() time.Time,
	sessionStore SessionStore,
	loginFailureStore LoginFailureStore,
) (*SessionManager, error) {
	if now == nil {
		now = time.Now
	}
	if sessionStore == nil {
		return nil, errors.New("auth session store is required")
	}
	if loginFailureStore == nil {
		return nil, errors.New("auth login failure store is required")
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
		accessTTL:         defaultAccessTokenTTL,
		refreshTTL:        defaultRefreshTokenTTL,
		now:               now,
		sessionStore:      sessionStore,
		loginFailureStore: loginFailureStore,
	}

	if err := manager.seedStaticAccessToken(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *SessionManager) PasswordPolicy() PasswordPolicy {
	return m.password
}

func (m *SessionManager) LockoutPolicy() LockoutPolicy {
	return m.lockout
}

func (m *SessionManager) Login(email string, password string) (Session, LoginFailure, bool) {
	session, failure, ok, _ := m.LoginWithError(email, password)
	return session, failure, ok
}

func (m *SessionManager) LoginWithError(email string, password string) (Session, LoginFailure, bool, error) {
	normalizedEmail := normalizeEmail(email)
	now := m.now().UTC()

	lockedUntil, locked, err := m.loginFailureStore.LockedUntil(normalizedEmail, now)
	if err != nil {
		return Session{}, LoginFailure{}, false, err
	}
	if locked {
		return Session{}, LoginFailure{
			Code:        LoginFailureLocked,
			Message:     "Account temporarily locked after repeated failed login attempts",
			LockedUntil: lockedUntil,
		}, false, nil
	}

	if failure := ValidatePasswordPolicy(password, m.password); failure != "" {
		if _, err := m.loginFailureStore.RecordFailure(normalizedEmail, now, m.lockout); err != nil {
			return Session{}, LoginFailure{}, false, err
		}
		return Session{}, LoginFailure{
			Code:    LoginFailurePasswordPolicy,
			Message: failure,
		}, false, nil
	}

	principal, ok := ValidateMockLogin(m.cfg, email, password)
	if !ok {
		lockedUntil, err := m.loginFailureStore.RecordFailure(normalizedEmail, now, m.lockout)
		if err != nil {
			return Session{}, LoginFailure{}, false, err
		}
		failure := LoginFailure{
			Code:    LoginFailureInvalidCredentials,
			Message: "Invalid email or password",
		}
		if !lockedUntil.IsZero() {
			failure.Code = LoginFailureLocked
			failure.Message = "Account temporarily locked after repeated failed login attempts"
			failure.LockedUntil = lockedUntil
		}
		return Session{}, failure, false, nil
	}

	session, err := m.issueSession(principal, now)
	if err != nil {
		return Session{}, LoginFailure{}, false, err
	}
	if err := m.loginFailureStore.Clear(normalizedEmail); err != nil {
		return Session{}, LoginFailure{}, false, err
	}
	return session, LoginFailure{}, true, nil
}

func (m *SessionManager) Refresh(refreshToken string) (Session, bool) {
	session, ok, _ := m.RefreshWithError(refreshToken)
	return session, ok
}

func (m *SessionManager) RefreshWithError(refreshToken string) (Session, bool, error) {
	now := m.now().UTC()
	return m.sessionStore.RotateRefreshToken(refreshToken, now, func(existing Session) Session {
		return m.newSession(existing.Principal, now)
	})
}

func (m *SessionManager) Logout(refreshToken string) bool {
	ok, _ := m.LogoutWithError(refreshToken)
	return ok
}

func (m *SessionManager) LogoutWithError(refreshToken string) (bool, error) {
	now := m.now().UTC()
	return m.sessionStore.RevokeRefreshToken(refreshToken, now)
}

func (m *SessionManager) AuthenticateAccessToken(accessToken string) (Principal, bool) {
	principal, ok, _ := m.AuthenticateAccessTokenWithError(accessToken)
	return principal, ok
}

func (m *SessionManager) AuthenticateAccessTokenWithError(accessToken string) (Principal, bool, error) {
	now := m.now().UTC()

	session, ok, err := m.sessionStore.FindByAccessToken(accessToken, now)
	if err != nil || !ok {
		return Principal{}, false, err
	}

	return session.Principal, true, nil
}

func (m *SessionManager) issueSession(principal Principal, now time.Time) (Session, error) {
	session := m.newSession(principal, now)
	if err := m.sessionStore.StoreSession(session, now); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (m *SessionManager) newSession(principal Principal, now time.Time) Session {
	return Session{
		AccessToken:      "local-at-" + randomToken(),
		RefreshToken:     "local-rt-" + randomToken(),
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshExpiresAt: now.Add(m.refreshTTL),
		Principal:        principal,
	}
}

func (m *SessionManager) seedStaticAccessToken() error {
	if strings.TrimSpace(m.cfg.AccessToken) == "" {
		return nil
	}

	now := m.now().UTC()
	session := Session{
		AccessToken:      m.cfg.AccessToken,
		RefreshToken:     "local-rt-static",
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshExpiresAt: now.Add(m.refreshTTL),
		Principal:        MockPrincipal(m.cfg),
	}

	return m.sessionStore.StoreSession(session, now)
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
