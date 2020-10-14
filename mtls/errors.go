package mtls

type Error string

func (e Error) Error() string { return string(e) }

const ClientCertificatesNotFoundError = Error("could not find client certificates")
const EmbeddedConfigDisabledError = Error("embedded config is disabled")
const HomeDirectoryError = Error("could not resolve user's home directory")
const MissingTLSConfigError = Error("missing required mTLS configuration")
const UnsupportedOSError = Error("running on unsupported operating system")
