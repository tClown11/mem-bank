# 智能决策引擎代码结构设计

**文档版本**: 1.0  
**创建日期**: 2025-08-26  
**作者**: AI记忆系统架构团队  
**状态**: 详细设计

## 1. 概述

本文档详细描述智能决策引擎在当前mem_bank项目中的代码结构设计，基于Clean Architecture原则，提供具体的文件组织、接口定义和实现方案。

## 2. 整体代码结构

### 2.1 目录结构设计

```
/mem_bank
├── internal/
│   ├── domain/
│   │   ├── memory/
│   │   │   ├── entity.go              # 现有记忆实体
│   │   │   ├── repository.go          # 现有存储接口
│   │   │   ├── service.go             # 现有服务接口
│   │   │   ├── intelligence.go        # 新增：智能处理接口
│   │   │   ├── pipeline.go            # 新增：管道接口定义
│   │   │   └── decision.go            # 新增：决策相关实体
│   │   └── user/ (现有)
│   │
│   ├── service/
│   │   ├── memory/
│   │   │   ├── service.go             # 现有基础服务
│   │   │   ├── ai_service.go          # 现有AI增强服务
│   │   │   ├── intelligent_service.go # 新增：智能决策服务
│   │   │   └── pipeline_service.go    # 新增：管道服务实现
│   │   ├── embedding/ (现有)
│   │   ├── intelligence/              # 新增：智能处理服务
│   │   │   ├── decision_engine.go     # 决策引擎实现
│   │   │   ├── extraction_stage.go    # 提取阶段实现
│   │   │   ├── decision_stage.go      # 决策阶段实现
│   │   │   ├── similarity_analyzer.go # 相似性分析器
│   │   │   └── context_builder.go     # 上下文构建器
│   │   └── llm/                       # 新增：LLM工具调用
│   │       ├── tool_manager.go        # 工具调用管理器
│   │       ├── tool_registry.go       # 工具注册中心
│   │       ├── handlers/              # 工具处理器
│   │       │   ├── memory_handlers.go # 记忆操作处理器
│   │       │   └── base_handler.go    # 基础处理器
│   │       └── prompt_builder.go      # 提示构建器
│   │
│   ├── handler/
│   │   ├── memory/
│   │   │   ├── handler.go            # 现有HTTP处理器
│   │   │   ├── intelligent_handler.go # 新增：智能处理处理器
│   │   │   └── job_handler.go        # 新增：作业状态处理器
│   │   └── user/ (现有)
│   │
│   ├── queue/
│   │   ├── interfaces.go             # 现有队列接口
│   │   ├── redis_queue.go            # 现有Redis队列
│   │   ├── memory_jobs.go            # 现有记忆作业
│   │   ├── intelligent_jobs.go       # 新增：智能处理作业
│   │   └── job_processor.go          # 新增：智能作业处理器
│   │
│   └── dao/ (现有目录保持不变)
│
├── pkg/
│   ├── llm/
│   │   ├── interfaces.go             # 现有LLM接口
│   │   ├── openai_provider.go        # 现有OpenAI实现
│   │   ├── tool_call.go              # 新增：工具调用支持
│   │   └── prompt_templates.go       # 新增：提示模板
│   ├── metrics/                      # 新增：指标收集
│   │   ├── intelligence_metrics.go   # 智能决策指标
│   │   └── collector.go              # 指标收集器
│   └── cache/                        # 新增：缓存优化
│       ├── ttl_cache.go              # TTL缓存实现
│       └── cache_manager.go          # 缓存管理器
│
├── configs/
│   ├── config.yaml                   # 现有配置文件
│   └── intelligence.yaml            # 新增：智能决策配置
│
└── docs/stage2/ (当前目录)
```

### 2.2 新增文件概览

| 文件路径 | 功能描述 | 优先级 |
|---------|---------|--------|
| `internal/domain/memory/intelligence.go` | 智能处理核心接口 | P0 |
| `internal/domain/memory/pipeline.go` | 记忆管道接口定义 | P0 |
| `internal/domain/memory/decision.go` | 决策相关实体定义 | P0 |
| `internal/service/intelligence/decision_engine.go` | 决策引擎核心实现 | P0 |
| `internal/service/llm/tool_manager.go` | LLM工具调用管理 | P0 |
| `internal/queue/intelligent_jobs.go` | 智能作业类型定义 | P1 |
| `pkg/llm/tool_call.go` | 工具调用框架 | P1 |
| `pkg/metrics/intelligence_metrics.go` | 智能决策监控指标 | P1 |

## 3. 核心接口和实体定义

### 3.1 智能处理接口 (`internal/domain/memory/intelligence.go`)

```go
package memory

import (
    "context"
    "time"
)

// IntelligentMemoryService 智能记忆服务接口
type IntelligentMemoryService interface {
    // ProcessIntelligentMemory 智能处理记忆输入
    ProcessIntelligentMemory(ctx context.Context, input *MemoryInput) (*ProcessingResult, error)
    
    // ProcessIntelligentMemoryAsync 异步智能处理
    ProcessIntelligentMemoryAsync(ctx context.Context, input *MemoryInput) (string, error)
    
    // GetProcessingStatus 获取处理状态
    GetProcessingStatus(ctx context.Context, jobID string) (*ProcessingStatus, error)
    
    // BatchProcessMemories 批量智能处理
    BatchProcessMemories(ctx context.Context, inputs []*MemoryInput) ([]*ProcessingResult, error)
}

// DecisionEngine 决策引擎接口
type DecisionEngine interface {
    // ProcessCandidateMemory 处理候选记忆
    ProcessCandidateMemory(ctx context.Context, userID string, candidate *CandidateMemory) (*MemoryOperation, error)
    
    // AnalyzeSimilarity 分析相似性
    AnalyzeSimilarity(ctx context.Context, candidate *CandidateMemory, existing []*Memory) (*SimilarityAnalysis, error)
    
    // MakeDecision 做出记忆操作决策
    MakeDecision(ctx context.Context, context *DecisionContext) (*MemoryOperation, error)
}

// MemoryPipeline 记忆处理管道接口
type MemoryPipeline interface {
    // ExtractCandidateMemories 提取候选记忆（阶段一）
    ExtractCandidateMemories(ctx context.Context, input *MemoryInput) ([]*CandidateMemory, error)
    
    // DecideMemoryOperations 决策记忆操作（阶段二）
    DecideMemoryOperations(ctx context.Context, userID string, candidates []*CandidateMemory) ([]*MemoryOperation, error)
    
    // ExecuteOperations 执行记忆操作
    ExecuteOperations(ctx context.Context, operations []*MemoryOperation) ([]*OperationResult, error)
}

// SimilarityAnalyzer 相似性分析器接口
type SimilarityAnalyzer interface {
    // Analyze 分析候选记忆与现有记忆的关系
    Analyze(candidate *CandidateMemory, existingMemories []*Memory) *SimilarityAnalysis
    
    // DetectConflicts 检测内容冲突
    DetectConflicts(candidate *CandidateMemory, existing *Memory) *ConflictInfo
    
    // CalculateConfidence 计算决策置信度
    CalculateConfidence(analysis *SimilarityAnalysis) float64
}
```

### 3.2 决策实体定义 (`internal/domain/memory/decision.go`)

```go
package memory

import (
    "encoding/json"
    "time"
)

// MemoryInput 记忆输入结构
type MemoryInput struct {
    UserID          string                 `json:"user_id" validate:"required,uuid"`
    SessionID       string                 `json:"session_id" validate:"required"`
    CurrentExchange ConversationExchange   `json:"current_exchange" validate:"required"`
    RecentHistory   []ConversationMessage  `json:"recent_history,omitempty"`
    RollingSummary  string                 `json:"rolling_summary,omitempty"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
    Timestamp       time.Time              `json:"timestamp"`
    Priority        int                    `json:"priority,omitempty"`
}

// ConversationExchange 对话交换
type ConversationExchange struct {
    UserMessage      string    `json:"user_message" validate:"required"`
    AssistantMessage string    `json:"assistant_message" validate:"required"`
    Timestamp        time.Time `json:"timestamp"`
}

// ConversationMessage 对话消息
type ConversationMessage struct {
    Role      string    `json:"role" validate:"required,oneof=user assistant"`
    Content   string    `json:"content" validate:"required"`
    Timestamp time.Time `json:"timestamp"`
}

// CandidateMemory 候选记忆
type CandidateMemory struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content" validate:"required,min=1,max=1000"`
    Type        MemoryType             `json:"type" validate:"required"`
    Importance  float64                `json:"importance" validate:"min=1,max=10"`
    Confidence  float64                `json:"confidence" validate:"min=0,max=1"`
    Entities    []string               `json:"entities,omitempty"`
    Keywords    []string               `json:"keywords,omitempty"`
    Reasoning   string                 `json:"reasoning,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    ExtractedAt time.Time              `json:"extracted_at"`
}

// MemoryType 记忆类型枚举
type MemoryType string

const (
    MemoryTypeFact         MemoryType = "fact"
    MemoryTypePreference   MemoryType = "preference"
    MemoryTypeGoal         MemoryType = "goal"
    MemoryTypeRelationship MemoryType = "relationship"
    MemoryTypeEvent        MemoryType = "event"
    MemoryTypeSkill        MemoryType = "skill"
)

// MemoryOperation 记忆操作
type MemoryOperation struct {
    ID            string                 `json:"id"`
    Type          OperationType          `json:"type" validate:"required"`
    TargetMemory  *CandidateMemory      `json:"target_memory,omitempty"`
    ExistingID    string                 `json:"existing_id,omitempty"`
    NewContent    string                 `json:"new_content,omitempty"`
    Reason        string                 `json:"reason" validate:"required"`
    Confidence    float64                `json:"confidence" validate:"min=0,max=1"`
    Priority      int                    `json:"priority,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt     time.Time              `json:"created_at"`
    ExecutedAt    *time.Time             `json:"executed_at,omitempty"`
    Status        OperationStatus        `json:"status"`
}

// OperationType 操作类型
type OperationType string

const (
    OperationAdd    OperationType = "ADD"
    OperationUpdate OperationType = "UPDATE"
    OperationDelete OperationType = "DELETE"
    OperationNoop   OperationType = "NOOP"
)

// OperationStatus 操作状态
type OperationStatus string

const (
    StatusPending   OperationStatus = "PENDING"
    StatusExecuted  OperationStatus = "EXECUTED"
    StatusFailed    OperationStatus = "FAILED"
    StatusSkipped   OperationStatus = "SKIPPED"
)

// SimilarityAnalysis 相似性分析结果
type SimilarityAnalysis struct {
    MaxSimilarity     float64           `json:"max_similarity"`
    HighlySimilar     []*Memory         `json:"highly_similar"`
    PotentialConflicts []*ConflictInfo  `json:"potential_conflicts"`
    Recommendations   []string          `json:"recommendations"`
    ConfidenceScore   float64           `json:"confidence_score"`
    AnalyzedAt        time.Time         `json:"analyzed_at"`
}

// ConflictInfo 冲突信息
type ConflictInfo struct {
    ConflictingMemory *Memory `json:"conflicting_memory"`
    ConflictType      string  `json:"conflict_type"`
    Severity          float64 `json:"severity" validate:"min=0,max=1"`
    Description       string  `json:"description"`
    DetectedAt        time.Time `json:"detected_at"`
}

// DecisionContext 决策上下文
type DecisionContext struct {
    Candidate        *CandidateMemory    `json:"candidate"`
    SimilarMemories  []*Memory           `json:"similar_memories"`
    Analysis         *SimilarityAnalysis `json:"analysis"`
    UserContext      *UserContext        `json:"user_context,omitempty"`
    DecisionCriteria *DecisionCriteria   `json:"decision_criteria"`
    Timestamp        time.Time           `json:"timestamp"`
}

// UserContext 用户上下文
type UserContext struct {
    UserID           string                 `json:"user_id"`
    Profile          *UserProfile           `json:"profile,omitempty"`
    RecentActivity   []*ActivityRecord      `json:"recent_activity,omitempty"`
    Preferences      map[string]interface{} `json:"preferences,omitempty"`
    SessionInfo      *SessionInfo           `json:"session_info,omitempty"`
}

// DecisionCriteria 决策标准
type DecisionCriteria struct {
    SimilarityThreshold   float64 `json:"similarity_threshold"`
    ImportanceThreshold   float64 `json:"importance_threshold"`
    ConfidenceThreshold   float64 `json:"confidence_threshold"`
    EnableConflictResolution bool `json:"enable_conflict_resolution"`
    MaxSimilarMemories    int     `json:"max_similar_memories"`
}

// ProcessingResult 处理结果
type ProcessingResult struct {
    JobID            string                 `json:"job_id"`
    Status           string                 `json:"status"`
    Message          string                 `json:"message"`
    CandidatesCount  int                    `json:"candidates_count"`
    OperationsCount  int                    `json:"operations_count"`
    ExecutionResults []*OperationResult     `json:"execution_results,omitempty"`
    Duration         time.Duration          `json:"duration"`
    Metadata         map[string]interface{} `json:"metadata,omitempty"`
    ProcessedAt      time.Time              `json:"processed_at"`
}

// OperationResult 操作结果
type OperationResult struct {
    OperationID string                 `json:"operation_id"`
    Success     bool                   `json:"success"`
    Message     string                 `json:"message,omitempty"`
    Error       string                 `json:"error,omitempty"`
    MemoryID    string                 `json:"memory_id,omitempty"`
    Duration    time.Duration          `json:"duration"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    ExecutedAt  time.Time              `json:"executed_at"`
}

// ProcessingStatus 处理状态
type ProcessingStatus struct {
    JobID       string                 `json:"job_id"`
    Status      string                 `json:"status"`
    Progress    float64                `json:"progress"`
    Message     string                 `json:"message,omitempty"`
    Error       string                 `json:"error,omitempty"`
    StartedAt   *time.Time             `json:"started_at,omitempty"`
    CompletedAt *time.Time             `json:"completed_at,omitempty"`
    Result      *ProcessingResult      `json:"result,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

## 4. 服务层实现

### 4.1 智能记忆服务 (`internal/service/memory/intelligent_service.go`)

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/queue"
    "mem_bank/pkg/metrics"
)

// IntelligentMemoryServiceImpl 智能记忆服务实现
type IntelligentMemoryServiceImpl struct {
    // 现有服务依赖
    memoryService    memory.Service
    embeddingService EmbeddingService
    jobQueue         queue.Queue
    
    // 新增智能组件
    decisionEngine   memory.DecisionEngine
    memoryPipeline   memory.MemoryPipeline
    
    // 基础设施
    config          *IntelligentConfig
    logger          *logrus.Logger
    metrics         *metrics.IntelligenceMetrics
}

// IntelligentConfig 智能服务配置
type IntelligentConfig struct {
    ExtractionConfig *ExtractionConfig `mapstructure:"extraction"`
    DecisionConfig   *DecisionConfig   `mapstructure:"decision"`
    ProcessingConfig *ProcessingConfig `mapstructure:"processing"`
}

// ExtractionConfig 提取配置
type ExtractionConfig struct {
    MaxCandidates       int     `mapstructure:"max_candidates"`
    MinImportance       float64 `mapstructure:"min_importance"`
    MinConfidence       float64 `mapstructure:"min_confidence"`
    EnableEntityExtraction bool `mapstructure:"enable_entity_extraction"`
    ContextWindowSize   int     `mapstructure:"context_window_size"`
}

// DecisionConfig 决策配置
type DecisionConfig struct {
    SimilarityThreshold    float64       `mapstructure:"similarity_threshold"`
    MaxSimilarMemories     int           `mapstructure:"max_similar_memories"`
    RequireHighConfidence  bool          `mapstructure:"require_high_confidence"`
    EnableConflictResolution bool        `mapstructure:"enable_conflict_resolution"`
    DecisionTimeout        time.Duration `mapstructure:"decision_timeout"`
}

// ProcessingConfig 处理配置
type ProcessingConfig struct {
    EnableAsync        bool          `mapstructure:"enable_async"`
    MaxRetries         int           `mapstructure:"max_retries"`
    RetryInterval      time.Duration `mapstructure:"retry_interval"`
    ProcessTimeout     time.Duration `mapstructure:"process_timeout"`
    BatchSize          int           `mapstructure:"batch_size"`
}

// NewIntelligentMemoryService 创建智能记忆服务
func NewIntelligentMemoryService(
    memoryService memory.Service,
    embeddingService EmbeddingService,
    jobQueue queue.Queue,
    decisionEngine memory.DecisionEngine,
    memoryPipeline memory.MemoryPipeline,
    config *IntelligentConfig,
    logger *logrus.Logger,
    metricsCollector *metrics.IntelligenceMetrics,
) memory.IntelligentMemoryService {
    return &IntelligentMemoryServiceImpl{
        memoryService:    memoryService,
        embeddingService: embeddingService,
        jobQueue:         jobQueue,
        decisionEngine:   decisionEngine,
        memoryPipeline:   memoryPipeline,
        config:          config,
        logger:          logger,
        metrics:         metricsCollector,
    }
}

// ProcessIntelligentMemory 智能处理记忆输入
func (ims *IntelligentMemoryServiceImpl) ProcessIntelligentMemory(ctx context.Context, input *memory.MemoryInput) (*memory.ProcessingResult, error) {
    startTime := time.Now()
    jobID := uuid.New().String()
    
    ims.logger.WithFields(logrus.Fields{
        "job_id":    jobID,
        "user_id":   input.UserID,
        "session_id": input.SessionID,
    }).Info("Starting intelligent memory processing")
    
    // 阶段一：提取候选记忆
    candidates, err := ims.memoryPipeline.ExtractCandidateMemories(ctx, input)
    if err != nil {
        ims.metrics.RecordProcessingError("extraction", err)
        return nil, fmt.Errorf("memory extraction failed: %w", err)
    }
    
    extractionDuration := time.Since(startTime)
    ims.metrics.RecordExtractionMetrics(input.UserID, len(candidates), extractionDuration, ims.extractMemoryTypes(candidates))
    
    if len(candidates) == 0 {
        return &memory.ProcessingResult{
            JobID:           jobID,
            Status:          "completed",
            Message:         "No valuable memories extracted",
            CandidatesCount: 0,
            OperationsCount: 0,
            Duration:        extractionDuration,
            ProcessedAt:     time.Now(),
        }, nil
    }
    
    // 阶段二：决策记忆操作
    decisionStartTime := time.Now()
    operations, err := ims.memoryPipeline.DecideMemoryOperations(ctx, input.UserID, candidates)
    if err != nil {
        ims.metrics.RecordProcessingError("decision", err)
        return nil, fmt.Errorf("memory decision failed: %w", err)
    }
    
    decisionDuration := time.Since(decisionStartTime)
    ims.metrics.RecordDecisionMetrics(operations, decisionDuration)
    
    // 执行操作
    executionResults, err := ims.memoryPipeline.ExecuteOperations(ctx, operations)
    if err != nil {
        ims.metrics.RecordProcessingError("execution", err)
        return nil, fmt.Errorf("operation execution failed: %w", err)
    }
    
    totalDuration := time.Since(startTime)
    
    result := &memory.ProcessingResult{
        JobID:            jobID,
        Status:           "completed",
        Message:          fmt.Sprintf("Successfully processed %d candidates, executed %d operations", len(candidates), len(operations)),
        CandidatesCount:  len(candidates),
        OperationsCount:  len(operations),
        ExecutionResults: executionResults,
        Duration:         totalDuration,
        ProcessedAt:      time.Now(),
        Metadata: map[string]interface{}{
            "extraction_duration": extractionDuration.String(),
            "decision_duration":   decisionDuration.String(),
            "candidate_types":     ims.extractMemoryTypes(candidates),
            "operation_types":     ims.extractOperationTypes(operations),
        },
    }
    
    ims.logger.WithFields(logrus.Fields{
        "job_id":          jobID,
        "user_id":         input.UserID,
        "candidates":      len(candidates),
        "operations":      len(operations),
        "duration":        totalDuration,
        "success_rate":    ims.calculateSuccessRate(executionResults),
    }).Info("Intelligent memory processing completed")
    
    ims.metrics.RecordTotalProcessingTime(totalDuration)
    
    return result, nil
}

// ProcessIntelligentMemoryAsync 异步智能处理
func (ims *IntelligentMemoryServiceImpl) ProcessIntelligentMemoryAsync(ctx context.Context, input *memory.MemoryInput) (string, error) {
    jobID := uuid.New().String()
    
    // 创建智能处理作业
    job := &queue.IntelligentMemoryJob{
        BaseJob: queue.BaseJob{
            ID:        jobID,
            Type:      queue.JobTypeIntelligentMemory,
            Status:    queue.JobStatusPending,
            CreatedAt: time.Now(),
            MaxRetries: ims.config.ProcessingConfig.MaxRetries,
        },
        Input: input,
        Config: &queue.IntelligentJobConfig{
            EnableExtraction:      true,
            EnableDecision:        true,
            ProcessingTimeout:     ims.config.ProcessingConfig.ProcessTimeout,
            ExtractionConfig:      ims.config.ExtractionConfig,
            DecisionConfig:        ims.config.DecisionConfig,
        },
    }
    
    // 提交到队列
    if err := ims.jobQueue.Enqueue(ctx, job); err != nil {
        ims.logger.WithError(err).WithField("job_id", jobID).Error("Failed to enqueue intelligent memory job")
        return "", fmt.Errorf("failed to queue intelligent memory processing: %w", err)
    }
    
    ims.logger.WithFields(logrus.Fields{
        "job_id":    jobID,
        "user_id":   input.UserID,
        "session_id": input.SessionID,
    }).Info("Intelligent memory processing job enqueued")
    
    return jobID, nil
}

// extractMemoryTypes 提取记忆类型用于指标
func (ims *IntelligentMemoryServiceImpl) extractMemoryTypes(candidates []*memory.CandidateMemory) []memory.MemoryType {
    types := make([]memory.MemoryType, len(candidates))
    for i, candidate := range candidates {
        types[i] = candidate.Type
    }
    return types
}

// extractOperationTypes 提取操作类型用于指标
func (ims *IntelligentMemoryServiceImpl) extractOperationTypes(operations []*memory.MemoryOperation) []memory.OperationType {
    types := make([]memory.OperationType, len(operations))
    for i, operation := range operations {
        types[i] = operation.Type
    }
    return types
}

// calculateSuccessRate 计算成功率
func (ims *IntelligentMemoryServiceImpl) calculateSuccessRate(results []*memory.OperationResult) float64 {
    if len(results) == 0 {
        return 0.0
    }
    
    successCount := 0
    for _, result := range results {
        if result.Success {
            successCount++
        }
    }
    
    return float64(successCount) / float64(len(results))
}
```

### 4.2 决策引擎实现 (`internal/service/intelligence/decision_engine.go`)

```go
package intelligence

import (
    "context"
    "fmt"
    "time"

    "github.com/sirupsen/logrus"

    "mem_bank/internal/domain/memory"
    "mem_bank/pkg/llm"
    "mem_bank/pkg/metrics"
)

// DecisionEngineImpl 决策引擎实现
type DecisionEngineImpl struct {
    llmProvider        llm.Provider
    memoryRepository   memory.Repository
    similarityAnalyzer memory.SimilarityAnalyzer
    promptBuilder      *PromptBuilder
    config            *DecisionConfig
    logger            *logrus.Logger
    metrics           *metrics.IntelligenceMetrics
}

// DecisionConfig 决策引擎配置
type DecisionConfig struct {
    SimilarityThreshold     float64       `mapstructure:"similarity_threshold"`
    MaxSimilarMemories      int           `mapstructure:"max_similar_memories"`
    RequireHighConfidence   bool          `mapstructure:"require_high_confidence"`
    EnableConflictResolution bool         `mapstructure:"enable_conflict_resolution"`
    DecisionTimeout         time.Duration `mapstructure:"decision_timeout"`
    MinConfidenceScore      float64       `mapstructure:"min_confidence_score"`
}

// NewDecisionEngine 创建决策引擎
func NewDecisionEngine(
    llmProvider llm.Provider,
    memoryRepository memory.Repository,
    similarityAnalyzer memory.SimilarityAnalyzer,
    promptBuilder *PromptBuilder,
    config *DecisionConfig,
    logger *logrus.Logger,
    metrics *metrics.IntelligenceMetrics,
) memory.DecisionEngine {
    return &DecisionEngineImpl{
        llmProvider:        llmProvider,
        memoryRepository:   memoryRepository,
        similarityAnalyzer: similarityAnalyzer,
        promptBuilder:      promptBuilder,
        config:            config,
        logger:            logger,
        metrics:           metrics,
    }
}

// ProcessCandidateMemory 处理候选记忆
func (de *DecisionEngineImpl) ProcessCandidateMemory(ctx context.Context, userID string, candidate *memory.CandidateMemory) (*memory.MemoryOperation, error) {
    startTime := time.Now()
    
    de.logger.WithFields(logrus.Fields{
        "user_id":     userID,
        "candidate_id": candidate.ID,
        "content_preview": de.truncateContent(candidate.Content, 50),
    }).Debug("Processing candidate memory")
    
    // 1. 生成候选记忆的向量嵌入
    embedding, err := de.generateEmbedding(ctx, candidate.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }
    
    // 2. 搜索相似的现有记忆
    similarMemories, err := de.memoryRepository.FindSimilar(ctx, &memory.SimilarityQuery{
        UserID:    userID,
        Embedding: embedding,
        Limit:     de.config.MaxSimilarMemories,
        Threshold: de.config.SimilarityThreshold,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to find similar memories: %w", err)
    }
    
    // 3. 分析相似性和冲突
    analysis := de.similarityAnalyzer.Analyze(candidate, similarMemories)
    
    // 4. 构建决策上下文并调用LLM
    decision, err := de.makeDecision(ctx, candidate, similarMemories, analysis)
    if err != nil {
        return nil, fmt.Errorf("failed to make decision: %w", err)
    }
    
    // 5. 验证决策质量
    if de.config.RequireHighConfidence && decision.Confidence < de.config.MinConfidenceScore {
        de.logger.WithFields(logrus.Fields{
            "candidate_id": candidate.ID,
            "confidence":   decision.Confidence,
            "min_required": de.config.MinConfidenceScore,
        }).Warn("Decision confidence below threshold, marking as NOOP")
        
        decision.Type = memory.OperationNoop
        decision.Reason = fmt.Sprintf("Decision confidence (%.2f) below required threshold (%.2f)", 
            decision.Confidence, de.config.MinConfidenceScore)
    }
    
    duration := time.Since(startTime)
    de.metrics.RecordDecisionLatency(duration)
    
    de.logger.WithFields(logrus.Fields{
        "user_id":         userID,
        "candidate_id":    candidate.ID,
        "decision_type":   decision.Type,
        "confidence":      decision.Confidence,
        "similar_count":   len(similarMemories),
        "duration":        duration,
    }).Info("Candidate memory decision completed")
    
    return decision, nil
}

// makeDecision 使用LLM做出决策
func (de *DecisionEngineImpl) makeDecision(ctx context.Context, candidate *memory.CandidateMemory, similarMemories []*memory.Memory, analysis *memory.SimilarityAnalysis) (*memory.MemoryOperation, error) {
    // 构建决策上下文
    decisionContext := &memory.DecisionContext{
        Candidate:       candidate,
        SimilarMemories: similarMemories,
        Analysis:        analysis,
        DecisionCriteria: &memory.DecisionCriteria{
            SimilarityThreshold:      de.config.SimilarityThreshold,
            ImportanceThreshold:      5.0, // 可配置
            ConfidenceThreshold:      de.config.MinConfidenceScore,
            EnableConflictResolution: de.config.EnableConflictResolution,
            MaxSimilarMemories:       de.config.MaxSimilarMemories,
        },
        Timestamp: time.Now(),
    }
    
    // 构建决策提示
    prompt, err := de.promptBuilder.BuildDecisionPrompt(decisionContext)
    if err != nil {
        return nil, fmt.Errorf("failed to build decision prompt: %w", err)
    }
    
    // 准备LLM请求
    request := &llm.ChatRequest{
        Messages: []llm.Message{{
            Role:    "user",
            Content: prompt,
        }},
        Tools:       de.getDecisionTools(),
        ToolChoice: &llm.ToolChoice{
            Type: "required",
        },
        Temperature: 0.2, // 低随机性，确保决策一致性
        MaxTokens:   1000,
        Timeout:     de.config.DecisionTimeout,
    }
    
    // 调用LLM进行决策
    response, err := de.llmProvider.ChatCompletion(ctx, request)
    if err != nil {
        de.metrics.RecordLLMError("decision", err)
        return nil, fmt.Errorf("LLM decision failed: %w", err)
    }
    
    de.metrics.RecordLLMCall("decision", response.Usage.TotalTokens)
    
    // 解析决策结果
    operation, err := de.parseDecisionResponse(response, candidate)
    if err != nil {
        return nil, fmt.Errorf("failed to parse decision response: %w", err)
    }
    
    return operation, nil
}

// generateEmbedding 生成嵌入向量
func (de *DecisionEngineImpl) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
    embeddings, err := de.llmProvider.CreateEmbeddings(ctx, []string{text})
    if err != nil {
        return nil, fmt.Errorf("failed to create embedding: %w", err)
    }
    
    if len(embeddings) == 0 {
        return nil, fmt.Errorf("no embedding generated")
    }
    
    return embeddings[0], nil
}

// getDecisionTools 获取决策工具集
func (de *DecisionEngineImpl) getDecisionTools() []llm.Tool {
    return []llm.Tool{
        {
            Name:        "add_new_memory",
            Description: "添加一个全新的记忆，当候选记忆包含之前未记录的重要信息时使用",
            Parameters: &llm.ToolParameters{
                Type: "object",
                Properties: map[string]*llm.ParameterProperty{
                    "content": {
                        Type:        "string",
                        Description: "要添加的记忆内容",
                    },
                    "importance": {
                        Type:        "number",
                        Description: "记忆重要性评分(1-10)",
                        Minimum:     1,
                        Maximum:     10,
                    },
                    "confidence": {
                        Type:        "number",
                        Description: "决策置信度(0.0-1.0)",
                        Minimum:     0,
                        Maximum:     1,
                    },
                    "reasoning": {
                        Type:        "string",
                        Description: "选择添加操作的详细理由，说明为什么这是新信息",
                    },
                    "tags": {
                        Type:        "array",
                        Description: "记忆标签，用于分类和检索",
                        Items: &llm.ParameterProperty{Type: "string"},
                    },
                },
                Required: []string{"content", "reasoning"},
            },
        },
        {
            Name:        "update_existing_memory",
            Description: "更新现有记忆，当候选记忆补充、修正或增强现有信息时使用",
            Parameters: &llm.ToolParameters{
                Type: "object",
                Properties: map[string]*llm.ParameterProperty{
                    "memory_id": {
                        Type:        "string",
                        Description: "要更新的现有记忆ID",
                    },
                    "new_content": {
                        Type:        "string",
                        Description: "更新后的记忆内容",
                    },
                    "merge_strategy": {
                        Type:        "string",
                        Description: "内容合并策略",
                        Enum:        []string{"replace", "append", "merge"},
                    },
                    "confidence": {
                        Type:        "number",
                        Description: "决策置信度(0.0-1.0)",
                        Minimum:     0,
                        Maximum:     1,
                    },
                    "reasoning": {
                        Type:        "string",
                        Description: "选择更新操作的详细理由，说明新旧信息的关系",
                    },
                },
                Required: []string{"memory_id", "new_content", "reasoning"},
            },
        },
        {
            Name:        "delete_memory",
            Description: "删除过时、错误或冲突的记忆，当候选记忆证明现有记忆不准确时使用",
            Parameters: &llm.ToolParameters{
                Type: "object",
                Properties: map[string]*llm.ParameterProperty{
                    "memory_id": {
                        Type:        "string",
                        Description: "要删除的记忆ID",
                    },
                    "replacement_content": {
                        Type:        "string",
                        Description: "替代的正确记忆内容（如果有的话）",
                    },
                    "confidence": {
                        Type:        "number",
                        Description: "决策置信度(0.0-1.0)",
                        Minimum:     0,
                        Maximum:     1,
                    },
                    "reasoning": {
                        Type:        "string",
                        Description: "选择删除操作的详细理由，说明为什么现有记忆需要删除",
                    },
                },
                Required: []string{"memory_id", "reasoning"},
            },
        },
        {
            Name:        "no_operation",
            Description: "不执行任何操作，当候选记忆重复、质量不足或不重要时使用",
            Parameters: &llm.ToolParameters{
                Type: "object",
                Properties: map[string]*llm.ParameterProperty{
                    "reasoning": {
                        Type:        "string",
                        Description: "不执行操作的详细理由，说明为什么候选记忆不值得保存",
                    },
                    "confidence": {
                        Type:        "number",
                        Description: "决策置信度(0.0-1.0)",
                        Minimum:     0,
                        Maximum:     1,
                    },
                },
                Required: []string{"reasoning"},
            },
        },
    }
}

// parseDecisionResponse 解析LLM决策响应
func (de *DecisionEngineImpl) parseDecisionResponse(response *llm.ChatResponse, candidate *memory.CandidateMemory) (*memory.MemoryOperation, error) {
    if len(response.Choices) == 0 {
        return nil, fmt.Errorf("no choices in LLM response")
    }
    
    choice := response.Choices[0]
    if len(choice.Message.ToolCalls) == 0 {
        return nil, fmt.Errorf("no tool calls in LLM response")
    }
    
    toolCall := choice.Message.ToolCalls[0]
    
    // 解析工具调用参数
    var params map[string]interface{}
    if err := toolCall.Function.Arguments.Unmarshal(&params); err != nil {
        return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
    }
    
    // 创建记忆操作
    operation := &memory.MemoryOperation{
        ID:           uuid.New().String(),
        TargetMemory: candidate,
        CreatedAt:    time.Now(),
        Status:       memory.StatusPending,
    }
    
    // 根据工具调用类型填充操作信息
    switch toolCall.Function.Name {
    case "add_new_memory":
        operation.Type = memory.OperationAdd
        operation.NewContent = de.getStringParam(params, "content")
        operation.Reason = de.getStringParam(params, "reasoning")
        operation.Confidence = de.getFloatParam(params, "confidence")
        if tags, ok := params["tags"].([]interface{}); ok {
            operation.Metadata = map[string]interface{}{"tags": tags}
        }
        
    case "update_existing_memory":
        operation.Type = memory.OperationUpdate
        operation.ExistingID = de.getStringParam(params, "memory_id")
        operation.NewContent = de.getStringParam(params, "new_content")
        operation.Reason = de.getStringParam(params, "reasoning")
        operation.Confidence = de.getFloatParam(params, "confidence")
        if strategy, ok := params["merge_strategy"].(string); ok {
            operation.Metadata = map[string]interface{}{"merge_strategy": strategy}
        }
        
    case "delete_memory":
        operation.Type = memory.OperationDelete
        operation.ExistingID = de.getStringParam(params, "memory_id")
        operation.Reason = de.getStringParam(params, "reasoning")
        operation.Confidence = de.getFloatParam(params, "confidence")
        if replacement, ok := params["replacement_content"].(string); ok && replacement != "" {
            operation.NewContent = replacement
        }
        
    case "no_operation":
        operation.Type = memory.OperationNoop
        operation.Reason = de.getStringParam(params, "reasoning")
        operation.Confidence = de.getFloatParam(params, "confidence")
        
    default:
        return nil, fmt.Errorf("unknown tool call: %s", toolCall.Function.Name)
    }
    
    // 验证必要字段
    if operation.Reason == "" {
        return nil, fmt.Errorf("missing reasoning in decision")
    }
    
    return operation, nil
}

// 辅助方法
func (de *DecisionEngineImpl) getStringParam(params map[string]interface{}, key string) string {
    if val, ok := params[key].(string); ok {
        return val
    }
    return ""
}

func (de *DecisionEngineImpl) getFloatParam(params map[string]interface{}, key string) float64 {
    if val, ok := params[key].(float64); ok {
        return val
    }
    return 0.0
}

func (de *DecisionEngineImpl) truncateContent(content string, maxLen int) string {
    if len(content) <= maxLen {
        return content
    }
    return content[:maxLen] + "..."
}
```

## 5. 队列系统扩展

### 5.1 智能作业定义 (`internal/queue/intelligent_jobs.go`)

```go
package queue

import (
    "encoding/json"
    "time"
    
    "mem_bank/internal/domain/memory"
)

// JobType 作业类型（扩展现有类型）
const (
    JobTypeIntelligentMemory JobType = "intelligent_memory"
    JobTypeBatchIntelligent  JobType = "batch_intelligent"
)

// IntelligentMemoryJob 智能记忆处理作业
type IntelligentMemoryJob struct {
    BaseJob
    Input  *memory.MemoryInput        `json:"input"`
    Config *IntelligentJobConfig      `json:"config,omitempty"`
    Result *memory.ProcessingResult   `json:"result,omitempty"`
}

// IntelligentJobConfig 智能作业配置
type IntelligentJobConfig struct {
    EnableExtraction    bool                                `json:"enable_extraction"`
    EnableDecision      bool                                `json:"enable_decision"`
    ProcessingTimeout   time.Duration                       `json:"processing_timeout"`
    ExtractionConfig    *intelligence.ExtractionConfig     `json:"extraction_config,omitempty"`
    DecisionConfig      *intelligence.DecisionConfig       `json:"decision_config,omitempty"`
}

// BatchIntelligentJob 批量智能处理作业
type BatchIntelligentJob struct {
    BaseJob
    Inputs    []*memory.MemoryInput     `json:"inputs"`
    Config    *IntelligentJobConfig     `json:"config,omitempty"`
    Results   []*memory.ProcessingResult `json:"results,omitempty"`
    BatchSize int                       `json:"batch_size"`
}

// MarshalJSON 实现JSON序列化
func (job *IntelligentMemoryJob) MarshalJSON() ([]byte, error) {
    type Alias IntelligentMemoryJob
    return json.Marshal(&struct {
        Type string `json:"type"`
        *Alias
    }{
        Type:  string(JobTypeIntelligentMemory),
        Alias: (*Alias)(job),
    })
}

// UnmarshalJSON 实现JSON反序列化
func (job *IntelligentMemoryJob) UnmarshalJSON(data []byte) error {
    type Alias IntelligentMemoryJob
    aux := &struct {
        Type string `json:"type"`
        *Alias
    }{
        Alias: (*Alias)(job),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    job.BaseJob.Type = JobType(aux.Type)
    return nil
}

// GetType 返回作业类型
func (job *IntelligentMemoryJob) GetType() JobType {
    return JobTypeIntelligentMemory
}

// GetTimeout 返回作业超时时间
func (job *IntelligentMemoryJob) GetTimeout() time.Duration {
    if job.Config != nil && job.Config.ProcessingTimeout > 0 {
        return job.Config.ProcessingTimeout
    }
    return 30 * time.Second // 默认超时时间
}
```

### 5.2 智能作业处理器 (`internal/queue/job_processor.go`)

```go
package queue

import (
    "context"
    "fmt"
    "time"
    
    "github.com/sirupsen/logrus"
    
    "mem_bank/internal/domain/memory"
    "mem_bank/pkg/metrics"
)

// IntelligentJobProcessor 智能作业处理器
type IntelligentJobProcessor struct {
    intelligentService memory.IntelligentMemoryService
    logger            *logrus.Logger
    metrics           *metrics.IntelligenceMetrics
}

// NewIntelligentJobProcessor 创建智能作业处理器
func NewIntelligentJobProcessor(
    intelligentService memory.IntelligentMemoryService,
    logger *logrus.Logger,
    metrics *metrics.IntelligenceMetrics,
) *IntelligentJobProcessor {
    return &IntelligentJobProcessor{
        intelligentService: intelligentService,
        logger:            logger,
        metrics:           metrics,
    }
}

// ProcessIntelligentMemoryJob 处理智能记忆作业
func (processor *IntelligentJobProcessor) ProcessIntelligentMemoryJob(ctx context.Context, job *IntelligentMemoryJob) error {
    startTime := time.Now()
    
    processor.logger.WithFields(logrus.Fields{
        "job_id":     job.ID,
        "user_id":    job.Input.UserID,
        "session_id": job.Input.SessionID,
        "type":       job.Type,
    }).Info("Processing intelligent memory job")
    
    // 更新作业状态为处理中
    job.Status = JobStatusProcessing
    job.StartedAt = &startTime
    
    // 处理智能记忆
    result, err := processor.intelligentService.ProcessIntelligentMemory(ctx, job.Input)
    if err != nil {
        job.Status = JobStatusFailed
        job.ErrorMessage = err.Error()
        processor.metrics.RecordJobFailure(string(job.Type), err)
        
        processor.logger.WithError(err).WithFields(logrus.Fields{
            "job_id":  job.ID,
            "user_id": job.Input.UserID,
        }).Error("Intelligent memory job processing failed")
        
        return fmt.Errorf("intelligent memory processing failed: %w", err)
    }
    
    // 更新作业结果和状态
    completedAt := time.Now()
    job.Result = result
    job.Status = JobStatusCompleted
    job.CompletedAt = &completedAt
    
    duration := completedAt.Sub(startTime)
    processor.metrics.RecordJobSuccess(string(job.Type), duration)
    
    processor.logger.WithFields(logrus.Fields{
        "job_id":          job.ID,
        "user_id":         job.Input.UserID,
        "candidates":      result.CandidatesCount,
        "operations":      result.OperationsCount,
        "duration":        duration,
    }).Info("Intelligent memory job processing completed successfully")
    
    return nil
}

// ProcessBatchIntelligentJob 处理批量智能作业
func (processor *IntelligentJobProcessor) ProcessBatchIntelligentJob(ctx context.Context, job *BatchIntelligentJob) error {
    startTime := time.Now()
    
    processor.logger.WithFields(logrus.Fields{
        "job_id":    job.ID,
        "batch_size": len(job.Inputs),
        "type":      job.Type,
    }).Info("Processing batch intelligent job")
    
    job.Status = JobStatusProcessing
    job.StartedAt = &startTime
    
    // 批量处理记忆输入
    results, err := processor.intelligentService.BatchProcessMemories(ctx, job.Inputs)
    if err != nil {
        job.Status = JobStatusFailed
        job.ErrorMessage = err.Error()
        processor.metrics.RecordJobFailure(string(job.Type), err)
        return fmt.Errorf("batch intelligent processing failed: %w", err)
    }
    
    completedAt := time.Now()
    job.Results = results
    job.Status = JobStatusCompleted
    job.CompletedAt = &completedAt
    
    duration := completedAt.Sub(startTime)
    processor.metrics.RecordJobSuccess(string(job.Type), duration)
    
    processor.logger.WithFields(logrus.Fields{
        "job_id":     job.ID,
        "batch_size": len(job.Inputs),
        "duration":   duration,
    }).Info("Batch intelligent job processing completed")
    
    return nil
}
```

## 6. API处理器扩展

### 6.1 智能处理API (`internal/handler/memory/intelligent_handler.go`)

```go
package memory

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/middleware"
)

// IntelligentMemoryHandler 智能记忆处理器
type IntelligentMemoryHandler struct {
    intelligentService memory.IntelligentMemoryService
    logger            *logrus.Logger
}

// NewIntelligentMemoryHandler 创建智能记忆处理器
func NewIntelligentMemoryHandler(
    intelligentService memory.IntelligentMemoryService,
    logger *logrus.Logger,
) *IntelligentMemoryHandler {
    return &IntelligentMemoryHandler{
        intelligentService: intelligentService,
        logger:            logger,
    }
}

// ProcessIntelligentMemoryRequest 智能记忆处理请求
type ProcessIntelligentMemoryRequest struct {
    UserID          string                       `json:"user_id" binding:"required,uuid"`
    SessionID       string                       `json:"session_id" binding:"required"`
    CurrentExchange memory.ConversationExchange `json:"current_exchange" binding:"required"`
    RecentHistory   []memory.ConversationMessage `json:"recent_history,omitempty"`
    RollingSummary  string                      `json:"rolling_summary,omitempty"`
    ProcessingMode  string                      `json:"processing_mode,omitempty"` // "sync" or "async"
    Priority        int                         `json:"priority,omitempty"`
}

// ProcessIntelligentMemory 智能处理记忆
// @Summary 智能处理记忆输入
// @Description 使用AI进行智能的记忆提取和决策处理
// @Tags Memory
// @Accept json
// @Produce json
// @Param request body ProcessIntelligentMemoryRequest true "智能处理请求"
// @Success 200 {object} memory.ProcessingResult "同步处理结果"
// @Success 202 {object} AsyncProcessingResponse "异步处理响应"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/memories/intelligent [post]
func (h *IntelligentMemoryHandler) ProcessIntelligentMemory(c *gin.Context) {
    var req ProcessIntelligentMemoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.WithError(err).Error("Invalid intelligent memory request")
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request parameters",
            "details": err.Error(),
        })
        return
    }
    
    // 构建记忆输入
    input := &memory.MemoryInput{
        UserID:          req.UserID,
        SessionID:       req.SessionID,
        CurrentExchange: req.CurrentExchange,
        RecentHistory:   req.RecentHistory,
        RollingSummary:  req.RollingSummary,
        Timestamp:       time.Now(),
        Priority:        req.Priority,
        Metadata: map[string]interface{}{
            "source":         "api",
            "processing_mode": req.ProcessingMode,
            "client_ip":      c.ClientIP(),
        },
    }
    
    // 确定处理模式
    processMode := req.ProcessingMode
    if processMode == "" {
        processMode = "async" // 默认异步处理
    }
    
    if processMode == "sync" {
        h.processSynchronous(c, input)
    } else {
        h.processAsynchronous(c, input)
    }
}

// processSynchronous 同步处理
func (h *IntelligentMemoryHandler) processSynchronous(c *gin.Context, input *memory.MemoryInput) {
    ctx := c.Request.Context()
    
    h.logger.WithFields(logrus.Fields{
        "user_id":    input.UserID,
        "session_id": input.SessionID,
        "mode":       "sync",
    }).Info("Processing intelligent memory synchronously")
    
    result, err := h.intelligentService.ProcessIntelligentMemory(ctx, input)
    if err != nil {
        h.logger.WithError(err).Error("Synchronous intelligent memory processing failed")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Intelligent memory processing failed",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// processAsynchronous 异步处理
func (h *IntelligentMemoryHandler) processAsynchronous(c *gin.Context, input *memory.MemoryInput) {
    ctx := c.Request.Context()
    
    h.logger.WithFields(logrus.Fields{
        "user_id":    input.UserID,
        "session_id": input.SessionID,
        "mode":       "async",
    }).Info("Processing intelligent memory asynchronously")
    
    jobID, err := h.intelligentService.ProcessIntelligentMemoryAsync(ctx, input)
    if err != nil {
        h.logger.WithError(err).Error("Asynchronous intelligent memory processing failed")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to queue intelligent memory processing",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusAccepted, AsyncProcessingResponse{
        JobID:   jobID,
        Status:  "queued",
        Message: "Intelligent memory processing has been queued",
        StatusURL: fmt.Sprintf("/api/v1/jobs/%s", jobID),
    })
}

// AsyncProcessingResponse 异步处理响应
type AsyncProcessingResponse struct {
    JobID     string `json:"job_id"`
    Status    string `json:"status"`
    Message   string `json:"message"`
    StatusURL string `json:"status_url"`
}

// BatchProcessIntelligentMemories 批量智能处理记忆
// @Summary 批量智能处理记忆
// @Description 批量处理多个记忆输入
// @Tags Memory
// @Accept json
// @Produce json
// @Param request body BatchProcessRequest true "批量处理请求"
// @Success 200 {array} memory.ProcessingResult "处理结果列表"
// @Success 202 {object} AsyncProcessingResponse "异步处理响应"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/memories/intelligent/batch [post]
func (h *IntelligentMemoryHandler) BatchProcessIntelligentMemories(c *gin.Context) {
    var req BatchProcessRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid batch request parameters",
            "details": err.Error(),
        })
        return
    }
    
    // 验证批次大小
    if len(req.Inputs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Empty batch request",
        })
        return
    }
    
    if len(req.Inputs) > 100 { // 限制批次大小
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Batch size too large",
            "max_allowed": 100,
            "requested": len(req.Inputs),
        })
        return
    }
    
    ctx := c.Request.Context()
    
    h.logger.WithFields(logrus.Fields{
        "batch_size": len(req.Inputs),
        "mode":      req.ProcessingMode,
    }).Info("Processing batch intelligent memories")
    
    if req.ProcessingMode == "sync" {
        // 同步批量处理
        results, err := h.intelligentService.BatchProcessMemories(ctx, req.Inputs)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Batch intelligent memory processing failed",
                "details": err.Error(),
            })
            return
        }
        
        c.JSON(http.StatusOK, BatchProcessResponse{
            Results: results,
            Summary: h.generateBatchSummary(results),
        })
    } else {
        // 异步批量处理（需要实现批量队列作业）
        c.JSON(http.StatusNotImplemented, gin.H{
            "error": "Async batch processing not yet implemented",
        })
    }
}

// BatchProcessRequest 批量处理请求
type BatchProcessRequest struct {
    Inputs         []*memory.MemoryInput `json:"inputs" binding:"required,min=1,max=100"`
    ProcessingMode string                `json:"processing_mode,omitempty"`
}

// BatchProcessResponse 批量处理响应
type BatchProcessResponse struct {
    Results []*memory.ProcessingResult `json:"results"`
    Summary BatchSummary              `json:"summary"`
}

// BatchSummary 批量处理摘要
type BatchSummary struct {
    TotalProcessed      int `json:"total_processed"`
    TotalCandidates     int `json:"total_candidates"`
    TotalOperations     int `json:"total_operations"`
    SuccessfulJobs      int `json:"successful_jobs"`
    FailedJobs          int `json:"failed_jobs"`
    AverageProcessingTime string `json:"average_processing_time"`
}

// generateBatchSummary 生成批量处理摘要
func (h *IntelligentMemoryHandler) generateBatchSummary(results []*memory.ProcessingResult) BatchSummary {
    summary := BatchSummary{
        TotalProcessed: len(results),
    }
    
    var totalDuration time.Duration
    for _, result := range results {
        summary.TotalCandidates += result.CandidatesCount
        summary.TotalOperations += result.OperationsCount
        totalDuration += result.Duration
        
        if result.Status == "completed" {
            summary.SuccessfulJobs++
        } else {
            summary.FailedJobs++
        }
    }
    
    if len(results) > 0 {
        avgDuration := totalDuration / time.Duration(len(results))
        summary.AverageProcessingTime = avgDuration.String()
    }
    
    return summary
}

// GetProcessingStatus 获取处理状态
// @Summary 获取智能处理状态
// @Description 获取异步智能处理的状态和结果
// @Tags Jobs
// @Produce json
// @Param job_id path string true "作业ID"
// @Success 200 {object} memory.ProcessingStatus "处理状态"
// @Failure 404 {object} ErrorResponse "作业未找到"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/jobs/{job_id} [get]
func (h *IntelligentMemoryHandler) GetProcessingStatus(c *gin.Context) {
    jobID := c.Param("job_id")
    if jobID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Missing job ID",
        })
        return
    }
    
    ctx := c.Request.Context()
    
    status, err := h.intelligentService.GetProcessingStatus(ctx, jobID)
    if err != nil {
        if err == memory.ErrJobNotFound {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Job not found",
                "job_id": jobID,
            })
            return
        }
        
        h.logger.WithError(err).WithField("job_id", jobID).Error("Failed to get job status")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get job status",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, status)
}
```

## 7. 配置管理扩展

### 7.1 智能决策配置 (`configs/intelligence.yaml`)

```yaml
# 智能决策引擎配置
intelligence:
  # 记忆提取配置
  extraction:
    max_candidates: 10
    min_importance: 5.0
    min_confidence: 0.6
    enable_entity_extraction: true
    context_window_size: 5
    
  # 决策配置  
  decision:
    similarity_threshold: 0.75
    max_similar_memories: 5
    require_high_confidence: true
    enable_conflict_resolution: true
    decision_timeout: 15s
    min_confidence_score: 0.7
    
  # 处理配置
  processing:
    enable_async: true
    max_retries: 3
    retry_interval: 5s
    process_timeout: 30s
    batch_size: 10

# LLM配置扩展
llm:
  provider: "openai"
  api_key: "${OPENAI_API_KEY}"
  base_url: "${OPENAI_BASE_URL:-https://api.openai.com/v1}"
  
  # 模型配置
  models:
    chat: "gpt-4-turbo-preview"
    embedding: "text-embedding-3-small"
    
  # 工具调用配置
  tool_calls:
    timeout: 20s
    max_retries: 2
    enable_parallel_calls: true
    
  # 缓存配置
  cache:
    enable_embedding_cache: true
    embedding_cache_ttl: 24h
    enable_response_cache: false
    response_cache_ttl: 1h

# 监控配置
metrics:
  enabled: true
  namespace: "mem_bank"
  subsystem: "intelligence"
  
  # 指标收集
  collect_extraction_metrics: true
  collect_decision_metrics: true
  collect_llm_metrics: true
  collect_cache_metrics: true
  
  # Prometheus配置
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"

# 缓存配置
cache:
  # 嵌入缓存
  embeddings:
    enabled: true
    ttl: 24h
    max_size: 10000
    
  # 相似性搜索缓存  
  similarity:
    enabled: true
    ttl: 1h
    max_size: 1000
    
  # 决策缓存
  decisions:
    enabled: false
    ttl: 30m
    max_size: 500
```

## 8. 测试结构

### 8.1 单元测试示例

```go
// internal/service/intelligence/decision_engine_test.go
package intelligence_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/service/intelligence"
    "mem_bank/pkg/llm/mocks"
)

func TestDecisionEngine_ProcessCandidateMemory(t *testing.T) {
    tests := []struct {
        name             string
        candidate        *memory.CandidateMemory
        similarMemories  []*memory.Memory
        expectedDecision memory.OperationType
        expectError      bool
    }{
        {
            name: "add_new_unique_memory",
            candidate: &memory.CandidateMemory{
                ID:         "candidate-1",
                Content:    "用户喜欢喝咖啡",
                Type:       memory.MemoryTypePreference,
                Importance: 7.0,
                Confidence: 0.9,
            },
            similarMemories:  []*memory.Memory{},
            expectedDecision: memory.OperationAdd,
            expectError:      false,
        },
        {
            name: "update_similar_memory",
            candidate: &memory.CandidateMemory{
                ID:         "candidate-2", 
                Content:    "用户现在更喜欢喝茶而不是咖啡",
                Type:       memory.MemoryTypePreference,
                Importance: 8.0,
                Confidence: 0.85,
            },
            similarMemories: []*memory.Memory{
                {
                    ID:      "memory-1",
                    Content: "用户喜欢喝咖啡",
                    UserID:  "user-1",
                },
            },
            expectedDecision: memory.OperationUpdate,
            expectError:      false,
        },
        {
            name: "no_operation_low_importance",
            candidate: &memory.CandidateMemory{
                ID:         "candidate-3",
                Content:    "今天天气不错",
                Type:       memory.MemoryTypeFact,
                Importance: 3.0,
                Confidence: 0.6,
            },
            similarMemories:  []*memory.Memory{},
            expectedDecision: memory.OperationNoop,
            expectError:      false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 设置模拟对象
            mockLLM := mocks.NewMockProvider(t)
            mockRepo := mocks.NewMockRepository(t)
            mockAnalyzer := mocks.NewMockSimilarityAnalyzer(t)
            
            // 配置模拟行为
            setupMockBehavior(mockLLM, mockRepo, mockAnalyzer, tt)
            
            // 创建决策引擎
            engine := intelligence.NewDecisionEngine(
                mockLLM,
                mockRepo,
                mockAnalyzer,
                &intelligence.PromptBuilder{},
                &intelligence.DecisionConfig{
                    SimilarityThreshold: 0.75,
                    MaxSimilarMemories:  5,
                    MinConfidenceScore:  0.7,
                },
                logger,
                metrics,
            )
            
            // 执行测试
            result, err := engine.ProcessCandidateMemory(context.Background(), "user-1", tt.candidate)
            
            // 验证结果
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.NotNil(t, result)
            assert.Equal(t, tt.expectedDecision, result.Type)
            assert.NotEmpty(t, result.Reason)
        })
    }
}
```

## 9. 部署和运维

### 9.1 Docker配置扩展

```dockerfile
# Dockerfile扩展
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用（包含智能决策功能）
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=intelligence-v1.0" \
    -o mem_bank_intelligent ./cmd/api/main.go

# 运行时镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制二进制文件和配置
COPY --from=builder /app/mem_bank_intelligent .
COPY ./configs/ ./configs/

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./mem_bank_intelligent"]
```

## 10. 总结

本代码结构设计为AI记忆银行系统的智能决策引擎提供了完整的实现框架：

### 10.1 核心特性
- **模块化设计**: 清晰的领域分层，便于维护和扩展
- **接口驱动**: 基于接口的依赖注入，支持测试和替换
- **配置灵活**: 完整的配置管理体系，支持不同环境
- **监控完备**: 全面的指标收集和监控支持
- **异步处理**: 高性能的异步作业处理系统

### 10.2 实施优先级
1. **P0**: 核心接口和决策引擎实现
2. **P1**: 智能作业处理和API接口
3. **P2**: 监控指标和缓存优化

### 10.3 预期效果
- 从基础存储升级为智能记忆管理
- 支持高并发的AI驱动决策处理
- 提供生产级的可观测性和运维能力
- 为企业级扩展奠定坚实基础

---

**文档状态**: 详细设计完成  
**实施准备**: 代码结构就绪  
**相关文档**: [智能决策引擎设计](./intelligent-decision-engine-design.md)