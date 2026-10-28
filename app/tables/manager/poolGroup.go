package manager

import (
	"app/poollog"
	"go.uber.org/zap"
)

// PoolGroup stores the applied definition. Changes are staged in PoolGroupOperation.
type PoolGroup struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentID      int64  `gorm:"column:agent_id;not null;index" json:"agentId"`
	Name         string `gorm:"size:100;not null" json:"name"`
	SourceSymbol string `gorm:"column:source_symbol;size:100;not null" json:"sourceSymbol"`
	Config       string `gorm:"type:longtext;not null" json:"-"`
	Version      int64  `gorm:"not null" json:"version"`
	Deleted      bool   `gorm:"not null;default:false" json:"deleted"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PoolGroup) TableName() (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("manager.PoolGroup.TableName", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return "pool_group"
}

type PoolGroupMember struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentID  int64  `gorm:"column:agent_id;not null;uniqueIndex:uq_pool_agent_game,priority:1" json:"agentId"`
	Symbol   string `gorm:"size:100;not null;uniqueIndex:uq_pool_agent_game,priority:2" json:"symbol"`
	GroupID  int64  `gorm:"column:group_id;not null;index" json:"groupId"`
	Position int    `gorm:"not null" json:"position"`
}

func (PoolGroupMember) TableName() (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("manager.PoolGroupMember.TableName", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return "pool_group_member"
}

// Request IDs are unique per agent. Payload contains the complete immutable intent.
type PoolGroupOperation struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentID    int64  `gorm:"column:agent_id;not null;uniqueIndex:uq_pool_operation_request,priority:1;index" json:"agentId"`
	RequestID  string `gorm:"column:request_id;size:80;not null;uniqueIndex:uq_pool_operation_request,priority:2" json:"requestId"`
	GroupID    int64  `gorm:"column:group_id;not null;index" json:"groupId"`
	Kind       string `gorm:"size:20;not null" json:"kind"`
	Payload    string `gorm:"type:longtext;not null" json:"-"`
	Before     string `gorm:"column:before_data;type:longtext" json:"-"`
	Status     string `gorm:"size:24;not null;index" json:"status"`
	Error      string `gorm:"type:text" json:"error"`
	OperatorID int64  `gorm:"column:operator_id;not null" json:"operatorId"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PoolGroupOperation) TableName() (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("manager.PoolGroupOperation.TableName", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return "pool_group_operation"
}
