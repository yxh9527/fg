package manager

// PoolScheme is a mutable pointer to immutable published revisions. Scope 0 means all sites.
type PoolScheme struct {
	ID         int64 `gorm:"primaryKey" json:"id"`
	WebID      int64 `gorm:"uniqueIndex;not null" json:"webId"`
	Version    int64 `json:"version"`
	RevisionID int64 `json:"revisionId"`
	UpdatedAt  int64 `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PoolSchemeRevision struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	SchemeID   int64  `gorm:"index" json:"schemeId"`
	Version    int64  `json:"version"`
	Groups     string `gorm:"type:longtext;not null" json:"-"`
	OperatorID int64  `json:"operatorId"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"createdAt"`
}

// The singleton registry serializes scope publication with agent registration.
type PoolSchemeRegistry struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false"`
}

type PoolAgentBinding struct {
	AgentID         int64  `gorm:"primaryKey;autoIncrement:false" json:"agentId"`
	Mode            string `gorm:"size:20;index;not null" json:"mode"`
	AppliedMode     string `gorm:"size:20" json:"appliedMode"`
	TargetRevision  int64  `json:"targetRevision"`
	AppliedRevision int64  `json:"appliedRevision"`
	Generation      int64  `json:"generation"`
	Initializing    bool   `json:"initializing"`
	Status          string `gorm:"size:24;index" json:"status"`
	Error           string `gorm:"type:text" json:"error"`
	UpdatedAt       int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PoolGroupMapping struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	AgentID    int64  `gorm:"uniqueIndex:uq_pool_group_mapping,priority:1;not null" json:"agentId"`
	LogicalKey string `gorm:"size:80;uniqueIndex:uq_pool_group_mapping,priority:2;not null" json:"key"`
	GroupID    int64  `gorm:"uniqueIndex;not null" json:"groupId"`
}

type PoolSyncBatch struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	RequestID    string `gorm:"size:80;uniqueIndex;not null" json:"requestId"`
	Fingerprint  string `gorm:"size:64;not null" json:"-"`
	Kind         string `gorm:"size:24;index;not null" json:"kind"`
	WebID        int64  `json:"webId"`
	RevisionID   int64  `json:"revisionId"`
	Cursor       int64  `json:"cursor"`
	MaxAgentID   int64  `json:"maxAgentId"`
	DispatchDone bool   `gorm:"index" json:"dispatchDone"`
	OperatorID   int64  `json:"operatorId"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PoolSyncItem struct {
	ID          int64  `gorm:"primaryKey" json:"id"`
	BatchID     int64  `gorm:"uniqueIndex:uq_pool_sync_agent,priority:1;not null;index" json:"batchId"`
	AgentID     int64  `gorm:"uniqueIndex:uq_pool_sync_agent,priority:2;not null;index" json:"agentId"`
	RevisionID  int64  `json:"revisionId"`
	Generation  int64  `json:"generation"`
	OperationID int64  `gorm:"index" json:"operationId"`
	Status      string `gorm:"size:24;index" json:"status"`
	Error       string `gorm:"type:text" json:"error"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}
