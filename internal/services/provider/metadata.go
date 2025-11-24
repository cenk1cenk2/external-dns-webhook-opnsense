package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificUUID    = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-uuid")
	ProviderSpecificDrifted = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-drifted")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
