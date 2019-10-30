// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/banzaicloud/istio-client-go/pkg/common/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VirtualService
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PolicySpec `json:"spec"`
}

// Policy defines what authentication methods can be accepted on workload(s),
// and if authenticated, which method/certificate will set the request principal
// (i.e request.auth.principal attribute).
//
// Authentication policy is composed of 2-part authentication:
// - peer: verify caller service credentials. This part will set source.user
// (peer identity).
// - origin: verify the origin credentials. This part will set request.auth.user
// (origin identity), as well as other attributes like request.auth.presenter,
// request.auth.audiences and raw claims. Note that the identity could be
// end-user, service account, device etc.
//
// Last but not least, the principal binding rule defines which identity (peer
// or origin) should be used as principal. By default, it uses peer.
//
// Examples:
//
// Policy to enable mTLS for all services in namespace frod. The policy name must be
// `default`, and it contains no rule for `targets`.
//
// ```yaml
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: default
//   namespace: frod
// spec:
//   peers:
//   - mtls:
// ```
// Policy to disable mTLS for "productpage" service
//
// ```yaml
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: productpage-mTLS-disable
//   namespace: frod
// spec:
//   targets:
//   - name: productpage
// ```
// Policy to require mTLS for peer authentication, and JWT for origin authentication
// for productpage:9000 except the path '/health_check' . Principal is set from origin identity.
//
// ```yaml
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: productpage-mTLS-with-JWT
//   namespace: frod
// spec:
//   targets:
//   - name: productpage
//     ports:
//     - number: 9000
//   peers:
//   - mtls:
//   origins:
//   - jwt:
//       issuer: "https://securetoken.google.com"
//       audiences:
//       - "productpage"
//       jwksUri: "https://www.googleapis.com/oauth2/v1/certs"
//       jwtHeaders:
//       - "x-goog-iap-jwt-assertion"
//       triggerRules:
//       - excludedPaths:
//         - exact: /health_check
//   principalBinding: USE_ORIGIN
// ```
type PolicySpec struct {
	// List rules to select workloads that the policy should be applied on.
	// If empty, policy will be used on all workloads in the same namespace.
	Targets []TargetSelector `json:"targets,omitempty"`

	// List of authentication methods that can be used for peer authentication.
	// They will be evaluated in order; the first validate one will be used to
	// set peer identity (source.user) and other peer attributes. If none of
	// these methods pass, request will be rejected with authentication failed error (401).
	// Leave the list empty if peer authentication is not required
	Peers []PeerAuthenticationMethod `json:"peers,omitempty"`

	// Set this flag to true to accept request (for peer authentication perspective),
	// even when none of the peer authentication methods defined above satisfied.
	// Typically, this is used to delay the rejection decision to next layer (e.g
	// authorization).
	// This flag is ignored if no authentication defined for peer (peers field is empty).
	PeerIsOptional bool `json:"peerIsOptional,omitempty"`

	// List of authentication methods that can be used for origin authentication.
	// Similar to peers, these will be evaluated in order; the first validate one
	// will be used to set origin identity and attributes (i.e request.auth.user,
	// request.auth.issuer etc). If none of these methods pass, request will be
	// rejected with authentication failed error (401).
	// A method may be skipped, depends on its trigger rule. If all of these methods
	// are skipped, origin authentication will be ignored, as if it is not defined.
	// Leave the list empty if origin authentication is not required.
	Origins []OriginAuthenticationMethod `json:"origins,omitempty"`

	// Set this flag to true to accept request (for origin authentication perspective),
	// even when none of the origin authentication methods defined above satisfied.
	// Typically, this is used to delay the rejection decision to next layer (e.g
	// authorization).
	// This flag is ignored if no authentication defined for origin (origins field is empty).
	OriginIsOptional bool `json:"originIsOptional,omitempty"`

	// Define whether peer or origin identity should be use for principal. Default
	// value is USE_PEER.
	// If peer (or origin) identity is not available, either because of peer/origin
	// authentication is not defined, or failed, principal will be left unset.
	// In other words, binding rule does not affect the decision to accept or
	// reject request.
	PrincipalBinding PrincipalBinding `json:"principalBinding,omitempty"`
}

// TargetSelector defines a matching rule to a workload. A workload is selected
// if it is associated with the service name and service port(s) specified in the selector rule.
type TargetSelector struct {
	// REQUIRED. The name must be a short name from the service registry. The
	// fully qualified domain name will be resolved in a platform specific manner.
	Name string `json:"name"`

	// Specifies the ports. Note that this is the port(s) exposed by the service, not workload instance ports.
	// For example, if a service is defined as below, then `8000` should be used, not `9000`.
	// ```yaml
	// kind: Service
	// metadata:
	//   ...
	// spec:
	//   ports:
	//   - name: http
	//     port: 8000
	//     targetPort: 9000
	//   selector:
	//     app: backend
	// ```
	// Leave empty to match all ports that are exposed.
	Ports []*PortSelector `json:"ports,omitempty"`
}

// PortSelector specifies the name or number of a port to be used for
// matching targets for authentication policy. This is copied from
// networking API to avoid dependency.
type PortSelector struct {
	// It is required to specify exactly one of these fields

	// Valid port number
	Number *uint32 `json:"number,omitempty"`

	// Port name
	Name *string `json:"name,omitempty"`
}

// PeerAuthenticationMethod defines one particular type of authentication, e.g
// mutual TLS, JWT etc, (no authentication is one type by itself) that can
// be used for peer authentication.
// The type can be progammatically determine by checking the type of the
// "params" field.
type PeerAuthenticationMethod struct {
	// It is required to specify exactly one of these fields

	// Set if mTLS is used.
	Mtls *MutualTLS `json:"mtls,omitempty"`

	// Set if JWT is used. This option is not yet available.
	Jwt *Jwt `json:"jwt,omitempty"`
}

// Defines the acceptable connection TLS mode.
type Mode string

const (
	// Client cert must be presented, connection is in TLS.
	ModeStrict Mode = "STRICT"

	// Connection can be either plaintext or TLS, and client cert can be omitted.
	ModePermissive Mode = "PERMISSIVE"
)

// TLS authentication params.
type MutualTLS struct {

	// WILL BE DEPRECATED, if set, will translates to `TLS_PERMISSIVE` mode.
	// Set this flag to true to allow regular TLS (i.e without client x509
	// certificate). If request carries client certificate, identity will be
	// extracted and used (set to peer identity). Otherwise, peer identity will
	// be left unset.
	// When the flag is false (default), request must have client certificate.
	AllowTLS bool `json:"allowTls,omitempty"`

	// Defines the mode of mTLS authentication.
	Mode Mode `json:"mode,omitempty"`
}

// JSON Web Token (JWT) token format for authentication as defined by
// [RFC 7519](https://tools.ietf.org/html/rfc7519). See [OAuth 2.0](https://tools.ietf.org/html/rfc6749) and
// [OIDC 1.0](http://openid.net/connect) for how this is used in the whole
// authentication flow.
//
// For example:
//
// A JWT for any requests:
//
// ```yaml
// issuer: https://example.com
// audiences:
// - bookstore_android.apps.googleusercontent.com
//   bookstore_web.apps.googleusercontent.com
// jwksUri: https://example.com/.well-known/jwks.json
// ```
//
// A JWT for all requests except request at path `/health_check` and path with
// prefix `/status/`. This is useful to expose some paths for public access but
// keep others JWT validated.
//
// ```yaml
// issuer: https://example.com
// jwksUri: https://example.com/.well-known/jwks.json
// triggerRules:
// - excludedPaths:
//   - exact: /health_check
//   - prefix: /status/
// ```
//
// A JWT only for requests at path `/admin`. This is useful to only require JWT
// validation on a specific set of paths but keep others public accessible.
//
// ```yaml
// issuer: https://example.com
// jwksUri: https://example.com/.well-known/jwks.json
// triggerRules:
// - includedPaths:
//   - prefix: /admin
// ```
//
// A JWT only for requests at path of prefix `/status/` but except the path of
// `/status/version`. This means for any request path with prefix `/status/` except
// `/status/version` will require a valid JWT to proceed.
//
// ```yaml
// issuer: https://example.com
// jwksUri: https://example.com/.well-known/jwks.json
// triggerRules:
// - excludedPaths:
//   - exact: /status/version
//   includedPaths:
//   - prefix: /status/
// ```
type Jwt struct {
	// Identifies the issuer that issued the JWT. See
	// [issuer](https://tools.ietf.org/html/rfc7519#section-4.1.1)
	// Usually a URL or an email address.
	//
	// Example: https://securetoken.google.com
	// Example: 1234567-compute@developer.gserviceaccount.com
	Issuer string `json:"issuer,omitempty"`

	// The list of JWT
	// [audiences](https://tools.ietf.org/html/rfc7519#section-4.1.3).
	// that are allowed to access. A JWT containing any of these
	// audiences will be accepted.
	//
	// The service name will be accepted if audiences is empty.
	//
	// Example:
	//
	// ```yaml
	// audiences:
	// - bookstore_android.apps.googleusercontent.com
	//   bookstore_web.apps.googleusercontent.com
	// ```
	Audiences []string `json:"audiences,omitempty"`

	// URL of the provider's public key set to validate signature of the
	// JWT. See [OpenID Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata).
	//
	// Optional if the key set document can either (a) be retrieved from
	// [OpenID
	// Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html) of
	// the issuer or (b) inferred from the email domain of the issuer (e.g. a
	// Google service account).
	//
	// Example: `https://www.googleapis.com/oauth2/v1/certs`
	//
	// Note: Only one of jwks_uri and jwks should be used.
	JwksURI string `json:"jwksUri,omitempty"`

	// JSON Web Key Set of public keys to validate signature of the JWT.
	// See https://auth0.com/docs/jwks.
	//
	// Note: Only one of jwks_uri and jwks should be used.
	Jwks string `json:"jwks,omitempty"`

	// Two fields below define where to extract the JWT from an HTTP request.
	//
	// If no explicit location is specified the following default
	// locations are tried in order:
	//
	//     1) The Authorization header using the Bearer schema,
	//        e.g. Authorization: Bearer <token>. (see
	//        [Authorization Request Header
	//        Field](https://tools.ietf.org/html/rfc6750#section-2.1))
	//
	//     2) `access_token` query parameter (see
	//     [URI Query Parameter](https://tools.ietf.org/html/rfc6750#section-2.3))

	// JWT is sent in a request header. `header` represents the
	// header name.
	//
	// For example, if `header=x-goog-iap-jwt-assertion`, the header
	// format will be `x-goog-iap-jwt-assertion: <JWT>`.
	JwtHeaders []string `json:"jwtHeaders,omitempty"`

	// JWT is sent in a query parameter. `query` represents the
	// query parameter name.
	//
	// For example, `query=jwt_token`.
	JwtParams []string `json:"jwtParams,omitempty"`

	// List of trigger rules to decide if this JWT should be used to validate the
	// request. The JWT validation happens if any one of the rules matched.
	// If the list is not empty and none of the rules matched, authentication will
	// skip the JWT validation.
	// Leave this empty to always trigger the JWT validation.
	TriggerRules []*TriggerRule `json:"triggerRules,omitempty"`
}

// Trigger rule to match against a request. The trigger rule is satisfied if
// and only if both rules, excluded_paths and include_paths are satisfied.
type TriggerRule struct {
	// List of paths to be excluded from the request. The rule is satisfied if
	// request path does not match to any of the path in this list.
	ExcludedPaths []v1alpha1.StringMatch `json:"excludedPaths,omitempty"`
	// List of paths that the request must include. If the list is not empty, the
	// rule is satisfied if request path matches at least one of the path in the list.
	// If the list is empty, the rule is ignored, in other words the rule is always satisfied.
	IncludedPaths []*v1alpha1.StringMatch `json:"includedPaths,omitempty"`
}

// OriginAuthenticationMethod defines authentication method/params for origin
// authentication. Origin could be end-user, device, delegate service etc.
// Currently, only JWT is supported for origin authentication.
type OriginAuthenticationMethod struct {
	// Jwt params for the method.
	Jwt *Jwt `json:"jwt,omitempty"`
}

// Associates authentication with request principal.
type PrincipalBinding string

const (
	// Principal will be set to the identity from peer authentication.
	PrincipalBindingUserPeer PrincipalBinding = "USE_PEER"
	// Principal will be set to the identity from origin authentication.
	PrincipalBindingUserOrigin PrincipalBinding = "USE_ORIGIN"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PolicyLIst is a list of Policy resources
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Policy `json:"items"`
}
