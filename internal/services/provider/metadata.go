package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificUUID        = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-uuid")
	ProviderSpecificDescription = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-description")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
