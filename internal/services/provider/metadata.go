package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificDescription = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-description")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
