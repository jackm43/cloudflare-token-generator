package cftoken

// ResourceScope indicates whether a service is scoped to a zone or account.
type ResourceScope string

const (
	ResourceScopeZone    ResourceScope = "zone"
	ResourceScopeAccount ResourceScope = "account"
)

// Permission maps a Cloudflare permission group name to its ID.
type Permission struct {
	ID   string
	Name string
}

// Service defines a Cloudflare service and the permissions needed to access it.
type Service struct {
	Name          string
	Description   string
	ResourceScope ResourceScope
	Permissions   []Permission
}

// Services maps service keys to their definitions.
var Services = map[string]Service{
	// Zone-scoped services
	"dns": {
		Name:          "dns",
		Description:   "DNS records management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "82e64a83756745bbbb1c9c2701bf816b", Name: "DNS Read"},
			{ID: "4755a26eedb94da69e1066d98aa820be", Name: "DNS Write"},
		},
	},
	"zone": {
		Name:          "zone",
		Description:   "Zone settings management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "c8fed203ed3043cba015a93ad1616f1f", Name: "Zone Read"},
			{ID: "3030687196b94b638145a3953da2b699", Name: "Zone Settings Write"},
		},
	},
	"cache": {
		Name:          "cache",
		Description:   "Cache purge",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "e17beae8b8cb423a99b1730f21238bed", Name: "Cache Purge"},
		},
	},
	"firewall": {
		Name:          "firewall",
		Description:   "Firewall services",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "4ec32dfcb35641c5bb32d5ef1ab963b4", Name: "Firewall Services Read"},
			{ID: "43137f8d07884d3198dc0ee77ca6e79b", Name: "Firewall Services Write"},
		},
	},
	"ssl": {
		Name:          "ssl",
		Description:   "SSL and certificates management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "7b7216b327b04b8fbc8f524e1f9b7531", Name: "SSL and Certificates Read"},
			{ID: "c03055bc037c4ea9afb9a9f104b7b721", Name: "SSL and Certificates Write"},
		},
	},
	"waf": {
		Name:          "waf",
		Description:   "Zone WAF management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "dbc512b354774852af2b5a5f4ba3d470", Name: "Zone WAF Read"},
			{ID: "fb6778dc191143babbfaa57993f1d275", Name: "Zone WAF Write"},
		},
	},
	"loadbalancer": {
		Name:          "loadbalancer",
		Description:   "Load balancer management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "e9a975f628014f1d85b723993116f7d5", Name: "Load Balancers Read"},
			{ID: "6d7f2f5f5b1d4a0e9081fdc98d432fd1", Name: "Load Balancers Write"},
		},
	},
	"pagerules": {
		Name:          "pagerules",
		Description:   "Page rules management",
		ResourceScope: ResourceScopeZone,
		Permissions: []Permission{
			{ID: "b415b70a4fd1412886f164451f20405c", Name: "Page Rules Read"},
			{ID: "ed07f6c337da4195b4e72a1fb2c6bcae", Name: "Page Rules Write"},
		},
	},

	// Account-scoped services
	"workers": {
		Name:          "workers",
		Description:   "Workers scripts management",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "1a71c399035b4950a1bd1466bbe4f420", Name: "Workers Scripts Read"},
			{ID: "e086da7e2179491d91ee5f35b3ca210a", Name: "Workers Scripts Write"},
		},
	},
	"kv": {
		Name:          "kv",
		Description:   "Workers KV storage",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "8b47d2786a534c08a1f94ee8f9f599ef", Name: "Workers KV Storage Read"},
			{ID: "f7f0eda5697f475c90846e879bab8666", Name: "Workers KV Storage Write"},
		},
	},
	"r2": {
		Name:          "r2",
		Description:   "Workers R2 object storage",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "b4992e1108244f5d8bfbd5744320c2e1", Name: "Workers R2 Storage Read"},
			{ID: "bf7481a1826f439697cb59a20b22293e", Name: "Workers R2 Storage Write"},
		},
	},
	"pages": {
		Name:          "pages",
		Description:   "Cloudflare Pages",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "e247aedd66bd41cc9193af0213416666", Name: "Pages Read"},
			{ID: "8d28297797f24fb8a29572aaac2f254d", Name: "Pages Write"},
		},
	},
	"d1": {
		Name:          "d1",
		Description:   "D1 database",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "192192df92ee43ac90f2aeeffce67e35", Name: "D1 Read"},
			{ID: "09b2857d1c31407795e75e3fed8617a1", Name: "D1 Write"},
		},
	},
	"queues": {
		Name:          "queues",
		Description:   "Cloudflare Queues",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "84a7755d54c646ca87cd50682a34bf7c", Name: "Queues Read"},
			{ID: "366f57075ffc42689627bcf8242a1b6d", Name: "Queues Write"},
		},
	},
	"ai": {
		Name:          "ai",
		Description:   "Workers AI inference",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "a92d2450e05d4e7bb7d0a64968f83d11", Name: "Workers AI Read"},
			{ID: "bacc64e0f6c34fc0883a1223f938a104", Name: "Workers AI Write"},
		},
	},
	"stream": {
		Name:          "stream",
		Description:   "Cloudflare Stream video",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "de21485a24744b76a004aa153898f7fe", Name: "Stream Read"},
			{ID: "714f9c13a5684c2885a793f5edb36f59", Name: "Stream Write"},
		},
	},
	"images": {
		Name:          "images",
		Description:   "Cloudflare Images",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "0cf6473ad41449e7b7b743d14fc20c60", Name: "Images Read"},
			{ID: "618ec6c64a3a42f8b08bdcb147ded4e4", Name: "Images Write"},
		},
	},
	"tunnels": {
		Name:          "tunnels",
		Description:   "Cloudflare Tunnel management",
		ResourceScope: ResourceScopeAccount,
		Permissions: []Permission{
			{ID: "efea2ab8357b47888938f101ae5e053f", Name: "Cloudflare Tunnel Read"},
			{ID: "c07321b023e944ff818fec44d8203567", Name: "Cloudflare Tunnel Write"},
		},
	},
}
