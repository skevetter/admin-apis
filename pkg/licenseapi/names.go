package licenseapi

type (
	ProductName    string
	ModuleName     string
	PlanStatus     string
	PlanInterval   string
	TierMode       string
	ResourceName   string
	ResourceStatus string
	TrialStatus    string
	FeatureStatus  string
	FeatureName    string
	ButtonName     string
)

// Metadata Keys
const (
	/* NEVER CHANGE ANY OF THESE */
	MetadataKeyAttachAll              = "attach_all_features"
	MetadataKeyProductForFeature      = "product_for_feature"
	MetadataKeyFeatureIsHidden        = "is_hidden"
	MetadataKeyFeatureIsPreview       = "is_preview"
	MetadataKeyFeatureIsLimit         = "is_limit"
	MetadataKeyFeatureLimitType       = "limit_type"
	MetadataKeyFeatureLimitTypeActive = "active"
	MetadataValueTrue                 = "true"
)

// Other
const (
	/* NEVER CHANGE ANY OF THESE */
	LimitsPrefix = "limits-"
)

// Products
const (
	/* NEVER CHANGE ANY OF THESE */
	Devsy     ProductName = "devsy"
	DevsyPro  ProductName = "devsy-pro"
	DevPodPro ProductName = "devpod-pro"
)

// Modules
const (
	KubernetesNamespaceModule ModuleName = "k8s-namespaces"
	KubernetesClusterModule   ModuleName = "k8s-clusters"
	DevsyModule               ModuleName = "devsy"
	DevsyProDistroModule      ModuleName = "devsy-pro-distro"
	DevPodModule              ModuleName = "devpod"
	AuthModule                ModuleName = "auth"
	TemplatingModule          ModuleName = "templating"
	SecretsModule             ModuleName = "secrets"
	DeploymentModesModule     ModuleName = "deployment-modes"
	UIModule                  ModuleName = "ui"
)

// Plan Status
const (
	PlanStatusActive    PlanStatus = "active"
	PlanStatusTrialing  PlanStatus = "trialing"
	PlanStatusLegacy    PlanStatus = "legacy"
	PlanStatusAvailable PlanStatus = ""
)

// Plan Interval
const (
	PlanIntervalMonth PlanInterval = "month"
	PlanIntervalYear  PlanInterval = "year"
)

// Tier Modes
const (
	TierModeGraduated TierMode = "graduated"
	TierModeVolume    TierMode = "volume"
)

// Resources (e.g. for limits)
const (
	/* NEVER CHANGE ANY OF THESE */
	ConnectedClusterLimit        ResourceName = "connected-cluster"
	DevsyClusterInstanceLimit    ResourceName = "virtual-cluster-instance"
	DevsyClusterInstanceHALimit  ResourceName = "virtual-cluster-instance-ha"
	SpaceInstanceLimit           ResourceName = "space-instance"
	DevPodWorkspaceInstanceLimit ResourceName = "devpod-workspace-instance"
	UserLimit                    ResourceName = "user"
	InstanceLimit                ResourceName = "instance"
)

// Resource Status
const (
	ResourceStatusActive       ResourceStatus = "active"
	ResourceStatusTotalCreated ResourceStatus = "created"
	ResourceStatusTotal        ResourceStatus = ""
)

// Trial Status
const (
	TrialStatusExpired TrialStatus = "expired"
	TrialStatusActive  TrialStatus = ""
)

// Buttons
const (
	ButtonContactSales  ButtonName = "contact-sales"
	ButtonManageBilling ButtonName = "manage-billing"
)

// Feature Status
const (
	FeatureStatusActive     FeatureStatus = "active"
	FeatureStatusPreview    FeatureStatus = "preview"
	FeatureStatusIncluded   FeatureStatus = "included"
	FeatureStatusHidden     FeatureStatus = "hidden"
	FeatureStatusDisallowed FeatureStatus = ""
)
