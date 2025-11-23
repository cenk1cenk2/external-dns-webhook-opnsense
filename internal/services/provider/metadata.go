package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificUUID    = ProviderSpecificMetadataKey("opnsense.record.uuid")
	ProviderSpecificDrifted = ProviderSpecificMetadataKey("opnsense.record.drifted")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}
