package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificUUID    = ProviderSpecificMetadataKey("uuid")
	ProviderSpecificDrifted = ProviderSpecificMetadataKey("drifted")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
