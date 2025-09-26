package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iden3/go-circuits/v2"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"
)

type jwzAuthMiddleware struct {
	verifier *auth.Verifier

	stateTransitionDelay time.Duration
	proofGenerationDelay time.Duration
	jwzGenerationDelay   time.Duration
}

// JWZAuthOption configures jwzAuthMiddleware optional parameters.
type JWZAuthOption func(*jwzAuthMiddleware)

// WithStateTransitionDelay sets a delay expected between state transitions and their verification.
func WithStateTransitionDelay(d time.Duration) JWZAuthOption {
	return func(m *jwzAuthMiddleware) {
		m.stateTransitionDelay = d
	}
}

// WithProofGenerationDelay sets an expected delay for proof generation.
func WithProofGenerationDelay(d time.Duration) JWZAuthOption {
	return func(m *jwzAuthMiddleware) {
		m.proofGenerationDelay = d
	}
}

// WithJWZGenerationDelay sets an expected delay for JWK generation.
func WithJWZGenerationDelay(d time.Duration) JWZAuthOption {
	return func(m *jwzAuthMiddleware) {
		m.jwzGenerationDelay = d
	}
}

// NewJWZAuthMiddleware creates a JWZ authentication middleware. Optional delays can be
// provided via functional options; callers may omit them.
func NewJWZAuthMiddleware(
	resolver map[string]pubsignals.StateResolver,
	opts ...JWZAuthOption,
) (func(http.Handler) http.Handler, error) {
	v, err := auth.NewVerifier(
		loaders.NewEmbeddedKeyLoader(),
		resolver,
	)
	if err != nil {
		return nil, err
	}
	m := &jwzAuthMiddleware{verifier: v}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	return m.JWZAuth, nil
}

func (v *jwzAuthMiddleware) getVerifyOpts() []pubsignals.VerifyOpt {
	opts := []pubsignals.VerifyOpt{
		pubsignals.WithAllowExpiredMessages(false),
	}
	if v.stateTransitionDelay != 0 {
		opts = append(opts,
			pubsignals.WithAcceptedStateTransitionDelay(v.stateTransitionDelay))
	}
	if v.proofGenerationDelay != 0 {
		opts = append(opts,
			pubsignals.WithAcceptedProofGenerationDelay(v.proofGenerationDelay))
	}
	return opts
}

func (v *jwzAuthMiddleware) JWZAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		jwzToken := authHeaderParts[1]

		t, err := v.verifier.VerifyJWZ(r.Context(),
			jwzToken, v.getVerifyOpts()...)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if err := verifyTokenExparation(t, v.jwzGenerationDelay); err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		authPubinputs := &circuits.AuthV2PubSignals{}
		if err := t.ParsePubSignals(authPubinputs); err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		fromDID, err := core.ParseDIDFromID(*authPubinputs.UserID)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := WithDIDContext(r.Context(), *fromDID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func verifyTokenExparation(t *jwz.Token, expirationDuration time.Duration) error {
	if expirationDuration == 0 {
		return nil
	}
	basicMessage := &iden3comm.BasicMessage{}
	if err := json.Unmarshal(t.GetPayload(), basicMessage); err != nil {
		return fmt.Errorf("failed to unmarshal basic message: %w", err)
	}

	if basicMessage.Type != protocol.AuthorizationResponseMessageType {
		return fmt.Errorf("invalid message type '%s' expected",
			protocol.AuthorizationResponseMessageType)
	}

	if basicMessage.CreatedTime == nil {
		return errors.New("missing created time in basic message")
	}

	createdAt := time.Unix(*basicMessage.CreatedTime, 0)
	tnow := time.Now().UTC()

	if createdAt.After(tnow.Add(5 * time.Minute)) {
		return errors.New("jwz has invalid created time in the future")
	}

	if createdAt.Add(expirationDuration).Before(tnow) {
		return errors.New("jwz has expired")
	}
	return nil
}
