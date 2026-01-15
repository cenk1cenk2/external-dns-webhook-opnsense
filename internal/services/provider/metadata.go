package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificUUID = ProviderSpecificMetadataKey("external-dns.alpha.kubernetes.io/opnsense-uuid")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
