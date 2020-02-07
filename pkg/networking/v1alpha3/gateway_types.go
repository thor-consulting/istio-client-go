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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// `Gateway` describes a load balancer operating at the edge of the mesh
// receiving incoming or outgoing HTTP/TCP connections. The specification
// describes a set of ports that should be exposed, the type of protocol to
// use, SNI configuration for the load balancer, etc.
//
// For example, the following Gateway configuration sets up a proxy to act
// as a load balancer exposing port 80 and 9080 (http), 443 (https),
// 9443(https) and port 2379 (TCP) for ingress.  The gateway will be
// applied to the proxy running on a pod with labels `app:
// my-gateway-controller`. While Istio will configure the proxy to listen
// on these ports, it is the responsibility of the user to ensure that
// external traffic to these ports are allowed into the mesh.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: Gateway
// metadata:
//   name: my-gateway
//   namespace: some-config-namespace
// spec:
//   selector:
//     app: my-gateway-controller
//   servers:
//   - port:
//       number: 80
//       name: http
//       protocol: HTTP
//     hosts:
//     - uk.bookinfo.com
//     - eu.bookinfo.com
//     tls:
//       httpsRedirect: true # sends 301 redirect for http requests
//   - port:
//       number: 443
//       name: https-443
//       protocol: HTTPS
//     hosts:
//     - uk.bookinfo.com
//     - eu.bookinfo.com
//     tls:
//       mode: SIMPLE # enables HTTPS on this port
//       serverCertificate: /etc/certs/servercert.pem
//       privateKey: /etc/certs/privatekey.pem
//   - port:
//       number: 9443
//       name: https-9443
//       protocol: HTTPS
//     hosts:
//     - "bookinfo-namespace/*.bookinfo.com"
//     tls:
//       mode: SIMPLE # enables HTTPS on this port
//       credentialName: bookinfo-secret # fetches certs from Kubernetes secret
//   - port:
//       number: 9080
//       name: http-wildcard
//       protocol: HTTP
//     hosts:
//     - "*"
//   - port:
//       number: 2379 # to expose internal service via external port 2379
//       name: mongo
//       protocol: MONGO
//     hosts:
//     - "*"
// ```
//
// The Gateway specification above describes the L4-L6 properties of a load
// balancer. A `VirtualService` can then be bound to a gateway to control
// the forwarding of traffic arriving at a particular host or gateway port.
//
// For example, the following VirtualService splits traffic for
// `https://uk.bookinfo.com/reviews`, `https://eu.bookinfo.com/reviews`,
// `http://uk.bookinfo.com:9080/reviews`,
// `http://eu.bookinfo.com:9080/reviews` into two versions (prod and qa) of
// an internal reviews service on port 9080. In addition, requests
// containing the cookie "user: dev-123" will be sent to special port 7777
// in the qa version. The same rule is also applicable inside the mesh for
// requests to the "reviews.prod.svc.cluster.local" service. This rule is
// applicable across ports 443, 9080. Note that `http://uk.bookinfo.com`
// gets redirected to `https://uk.bookinfo.com` (i.e. 80 redirects to 443).
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: bookinfo-rule
//   namespace: bookinfo-namespace
// spec:
//   hosts:
//   - reviews.prod.svc.cluster.local
//   - uk.bookinfo.com
//   - eu.bookinfo.com
//   gateways:
//   - some-config-namespace/my-gateway
//   - mesh # applies to all the sidecars in the mesh
//   http:
//   - match:
//     - headers:
//         cookie:
//           exact: "user=dev-123"
//     route:
//     - destination:
//         port:
//           number: 7777
//         host: reviews.qa.svc.cluster.local
//   - match:
//     - uri:
//         prefix: /reviews/
//     route:
//     - destination:
//         port:
//           number: 9080 # can be omitted if it's the only port for reviews
//         host: reviews.prod.svc.cluster.local
//       weight: 80
//     - destination:
//         host: reviews.qa.svc.cluster.local
//       weight: 20
// ```
//
// The following VirtualService forwards traffic arriving at (external)
// port 27017 to internal Mongo server on port 5555. This rule is not
// applicable internally in the mesh as the gateway list omits the
// reserved name `mesh`.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: bookinfo-Mongo
//   namespace: bookinfo-namespace
// spec:
//   hosts:
//   - mongosvr.prod.svc.cluster.local # name of internal Mongo service
//   gateways:
//   - some-config-namespace/my-gateway # can omit the namespace if gateway is in same
//                                        namespace as virtual service.
//   tcp:
//   - match:
//     - port: 27017
//     route:
//     - destination:
//         host: mongo.prod.svc.cluster.local
//         port:
//           number: 5555
// ```
//
// It is possible to restrict the set of virtual services that can bind to
// a gateway server using the namespace/hostname syntax in the hosts field.
// For example, the following Gateway allows any virtual service in the ns1
// namespace to bind to it, while restricting only the virtual service with
// foo.bar.com host in the ns2 namespace to bind to it.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: Gateway
// metadata:
//   name: my-gateway
//   namespace: some-config-namespace
// spec:
//   selector:
//     app: my-gateway-controller
//   servers:
//   - port:
//       number: 80
//       name: http
//       protocol: HTTP
//     hosts:
//     - "ns1/*"
//     - "ns2/foo.bar.com"
// ```
//
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	// REQUIRED: A list of server specifications.
	Servers []Server `json:"servers"`

	// REQUIRED: One or more labels that indicate a specific set of pods/VMs
	// on which this gateway configuration should be applied. The scope of
	// label search is restricted to the configuration namespace in which the
	// the resource is present. In other words, the Gateway resource must
	// reside in the same namespace as the gateway workload instance.
	Selector map[string]string `json:"selector,omitempty"`
}

// `Server` describes the properties of the proxy on a given load balancer
// port. For example,
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: Gateway
// metadata:
//   name: my-ingress
// spec:
//   selector:
//     app: my-ingress-gateway
//   servers:
//   - port:
//       number: 80
//       name: http2
//       protocol: HTTP2
//     hosts:
//     - "*"
// ```
//
// Another example
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: Gateway
// metadata:
//   name: my-tcp-ingress
// spec:
//   selector:
//     app: my-tcp-ingress-gateway
//   servers:
//   - port:
//       number: 27018
//       name: mongo
//       protocol: MONGO
//     hosts:
//     - "*"
// ```
//
// The following is an example of TLS configuration for port 443
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: Gateway
// metadata:
//   name: my-tls-ingress
// spec:
//   selector:
//     app: my-tls-ingress-gateway
//   servers:
//   - port:
//       number: 443
//       name: https
//       protocol: HTTPS
//     hosts:
//     - "*"
//     tls:
//       mode: SIMPLE
//       serverCertificate: /etc/certs/server.pem
//       privateKey: /etc/certs/privatekey.pem
// ```
type Server struct {
	// REQUIRED: The Port on which the proxy should listen for incoming
	// connections.
	Port *Port `json:"port"`

	// REQUIRED. One or more hosts exposed by this gateway.
	// While typically applicable to
	// HTTP services, it can also be used for TCP services using TLS with SNI.
	// A host is specified as a `dnsName` with an optional `namespace/` prefix.
	// The `dnsName` should be specified using FQDN format, optionally including
	// a wildcard character in the left-most component (e.g., `prod/*.example.com`).
	// Set the `dnsName` to `*` to select all `VirtualService` hosts from the
	// specified namespace (e.g.,`prod/*`).
	//
	// The `namespace` can be set to `*` or `.`, representing any or the current
	// namespace, respectively. For example, `*/foo.example.com` selects the
	// service from any available namespace while `./foo.example.com` only selects
	// the service from the namespace of the sidecar. The default, if no `namespace/`
	// is specified, is `*/`, that is, select services from any namespace.
	// Any associated `DestinationRule` in the selected namespace will also be used.
	//
	// A `VirtualService` must be bound to the gateway and must have one or
	// more hosts that match the hosts specified in a server. The match
	// could be an exact match or a suffix match with the server's hosts. For
	// example, if the server's hosts specifies `*.example.com`, a
	// `VirtualService` with hosts `dev.example.com` or `prod.example.com` will
	// match. However, a `VirtualService` with host `example.com` or
	// `newexample.com` will not match.
	//
	// NOTE: Only virtual services exported to the gateway's namespace
	// (e.g., `exportTo` value of `*`) can be referenced.
	// Private configurations (e.g., `exportTo` set to `.`) will not be
	// available. Refer to the `exportTo` setting in `VirtualService`,
	// `DestinationRule`, and `ServiceEntry` configurations for details.
	Hosts []string `json:"hosts,omitempty"`

	// Set of TLS related options that govern the server's behavior. Use
	// these options to control if all http requests should be redirected to
	// https, and the TLS modes to use.
	TLS *TLSOptions `json:"tls,omitempty"`

	// The loopback IP endpoint or Unix domain socket to which traffic should
	// be forwarded to by default. Format should be `127.0.0.1:PORT` or
	// `unix:///path/to/socket` or `unix://@foobar` (Linux abstract namespace).
	DefaultEndpoint *string `json:"defaultEndpoint,omitempty"`
}

type TLSOptions struct {
	// If set to true, the load balancer will send a 301 redirect for all
	// http connections, asking the clients to use HTTPS.
	HTTPSRedirect *bool `json:"httpsRedirect,omitempty"`

	// Optional: Indicates whether connections to this port should be
	// secured using TLS. The value of this field determines how TLS is
	// enforced.
	Mode TLSMode `json:"mode,omitempty"`

	// REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file
	// holding the server-side TLS certificate to use.
	ServerCertificate *string `json:"serverCertificate,omitempty"`

	// REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file
	// holding the server's private key.
	PrivateKey *string `json:"privateKey,omitempty"`

	// REQUIRED if mode is `MUTUAL`. The path to a file containing
	// certificate authority certificates to use in verifying a presented
	// client side certificate.
	CaCertificates *string `json:"caCertificates,omitempty"`

	// The credentialName stands for a unique identifier that can be used
	// to identify the serverCertificate and the privateKey. The
	// credentialName appended with suffix "-cacert" is used to identify
	// the CaCertificates associated with this server. Gateway workloads
	// capable of fetching credentials from a remote credential store such
	// as Kubernetes secrets, will be configured to retrieve the
	// serverCertificate and the privateKey using credentialName, instead
	// of using the file system paths specified above. If using mutual TLS,
	// gateway workload instances will retrieve the CaCertificates using
	// credentialName-cacert. The semantics of the name are platform
	// dependent.  In Kubernetes, the default Istio supplied credential
	// server expects the credentialName to match the name of the
	// Kubernetes secret that holds the server certificate, the private
	// key, and the CA certificate (if using mutual TLS). Set the
	// `ISTIO_META_USER_SDS` metadata variable in the gateway's proxy to
	// enable the dynamic credential fetching feature.
	CredentialName *string `json:"credentialName,omitempty"`

	// A list of alternate names to verify the subject identity in the
	// certificate presented by the client.
	SubjectAltNames []string `json:"subjectAltNames,omitempty"`

	// An optional list of base64-encoded SHA-256 hashes of the SKPIs of
	// authorized client certificates.
	// Note: When both verify_certificate_hash and verify_certificate_spki
	// are specified, a hash matching either value will result in the
	// certificate being accepted.
	VerifyCertificateSpki []string `json:"verifyCertificateSpki,omitempty"`

	// An optional list of hex-encoded SHA-256 hashes of the
	// authorized client certificates. Both simple and colon separated
	// formats are acceptable.
	// Note: When both verify_certificate_hash and verify_certificate_spki
	// are specified, a hash matching either value will result in the
	// certificate being accepted.
	VerifyCertificateHash []string `json:"verifyCertificateHash,omitempty"`

	// Optional: Minimum TLS protocol version.
	MinProtocolVersion *TLSProtocol `json:"minProtocolVersion,omitempty"`

	// Optional: Maximum TLS protocol version.
	MaxProtocolVersion *TLSProtocol `json:"maxProtocolVersion,omitempty"`

	// Optional: If specified, only support the specified cipher list.
	// Otherwise default to the default cipher list supported by Envoy.
	CipherSuites []string `json:"cipherSuites,omitempty"`
}

// TLS protocol versions.
type TLSProtocol string

const (
	// Automatically choose the optimal TLS version.
	TLSProtocolAuto TLSProtocol = "TLS_AUTO"

	// TLS version 1.0
	TLSProtocolV10 TLSProtocol = "TLSV1_0"

	// TLS version 1.1
	TLSProtocolV11 TLSProtocol = "TLSV1_1"

	// TLS version 1.2
	TLSProtocolV12 TLSProtocol = "TLSV1_2"

	// TLS version 1.3
	TLSProtocolV13 TLSProtocol = "TLSV1_3"
)

// TLS modes enforced by the proxy
type TLSMode string

const (
	// The SNI string presented by the client will be used as the match
	// criterion in a VirtualService TLS route to determine the
	// destination service from the service registry.
	TLSModePassThrough TLSMode = "PASSTHROUGH"

	// Secure connections with standard TLS semantics.
	TLSModeSimple TLSMode = "SIMPLE"

	// Secure connections to the downstream using mutual TLS by presenting
	// server certificates for authentication.
	TLSModeMutual TLSMode = "MUTUAL"

	// Similar to the passthrough mode, except servers with this TLS mode
	// do not require an associated VirtualService to map from the SNI
	// value to service in the registry. The destination details such as
	// the service/subset/port are encoded in the SNI value. The proxy
	// will forward to the upstream (Envoy) cluster (a group of
	// endpoints) specified by the SNI value. This server is typically
	// used to provide connectivity between services in disparate L3
	// networks that otherwise do not have direct connectivity between
	// their respective endpoints. Use of this mode assumes that both the
	// source and the destination are using Istio mTLS to secure traffic.
	TLSModeMutualAutoPassThrough TLSMode = "AUTO_PASSTHROUGH"

	// Secure connections from the downstream using mutual TLS by presenting
	// server certificates for authentication.
	// Compared to Mutual mode, this mode uses certificates, representing
	// gateway workload identity, generated automatically by Istio for
	// mTLS authentication. When this mode is used, all other fields in
	// `TLSOptions` should be empty.
	TLSModeIstioMutual TLSMode = "ISTIO_MUTUAL"
)

// Port describes the properties of a specific port of a service.
type Port struct {
	// REQUIRED: A valid non-negative integer port number.
	Number int `json:"number"`

	// REQUIRED: The protocol exposed on the port.
	// MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS.
	// TLS implies the connection will be routed based on the SNI header to
	// the destination without terminating the TLS connection.
	Protocol PortProtocol `json:"protocol"`

	// Label assigned to the port.
	Name string `json:"name,omitempty"`
}

type PortProtocol string

const (
	ProtocolHTTP    PortProtocol = "HTTP"
	ProtocolHTTPS   PortProtocol = "HTTPS"
	ProtocolGRPC    PortProtocol = "GRPC"
	ProtocolGRPCWeb PortProtocol = "GRPC-Web"
	ProtocolHTTP2   PortProtocol = "HTTP2"
	ProtocolMongo   PortProtocol = "Mongo"
	ProtocolTCP     PortProtocol = "TCP"
	ProtocolTLS     PortProtocol = "TLS"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayList is a list of Gateway resources
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Gateway `json:"items"`
}
