package provider

type ProviderSpecificMetadataKey string

const (
	ProviderSpecificDescription = ProviderSpecificMetadataKey("external-dns.kubernetes.io/opnsense-description")
)

func (k ProviderSpecificMetadataKey) String() string {
	return string(k)
}

type EndpointLabel string

const (
	EndpointLabelSetIdentifier = EndpointLabel("set-identifier")
	EndpointLabelUUID          = EndpointLabel("uuid")
)

func (l EndpointLabel) String() string {
	return string(l)
}
