# ReAct Agent - Refactored Implementation

This directory contains a comprehensive refactored implementation of the ReAct (Reasoning and Acting) agent, consolidating all functionality into a single `agent` package with improved architecture, error handling, and monitoring capabilities.

## Architecture Overview

The refactored agent follows the official `eino` implementation patterns and provides a complete ReAct agent with the following key improvements:

1. **Consolidated Structure**: All files are now in the `agent` directory
2. **Eliminated Duplicates**: Removed duplicate data structures like `ExecutionStep`
3. **Enhanced Flow Implementation**: Proper tool calling and graph orchestration
4. **Official eino Integration**: Based on the official eino ReAct implementation

## File Structure

```
agent/
├── react_agent.go      # Main agent implementation with eino graph
├── config.go           # Configuration management
├── reasoning.go        # Reasoning engine for decision making
├── decision.go         # Decision controller for action determination
├── tools.go           # Tool management and execution
├── state.go           # State management and tracking
├── monitor.go         # Monitoring and debugging capabilities
├── errors.go          # Comprehensive error handling
├── example.go         # Usage examples and demonstrations
└── README.md          # This documentation
```

## Key Components

### 1. Agent (`react_agent.go`)
The main agent implementation that orchestrates all components using an `eino` graph:
- **Graph-based execution**: Uses `eino.compose.Graph` for proper orchestration
- **Node-based processing**: Separate nodes for reasoning, decision, tool execution
- **Conditional routing**: Smart routing based on decisions and state
- **Error recovery**: Built-in retry mechanisms and error handling

### 2. Configuration (`config.go`)
Comprehensive configuration management:
- Model configuration
- Tool setup
- Execution parameters
- Debug settings
- Graph options

### 3. Reasoning Engine (`reasoning.go`)
Handles the agent's reasoning process:
- Prompt building
- Response parsing (JSON and text formats)
- Reasoning validation
- Context management

### 4. Decision Controller (`decision.go`)
Manages decision-making logic:
- Action determination (tool call, final answer, continue)
- Decision validation
- Quality assessment
- Error handling

### 5. Tool Manager (`tools.go`)
Handles tool execution and management:
- Tool discovery and registration
- Argument processing
- Execution monitoring
- Result formatting

### 6. State Manager (`state.go`)
Manages agent state throughout execution:
- State initialization and updates
- Execution trace creation
- Message management
- History tracking

### 7. Monitor (`monitor.go`)
Provides comprehensive monitoring and debugging:
- Execution metrics
- Performance tracking
- Debug logging
- Success rate analysis

### 8. Error Handling (`errors.go`)
Structured error management:
- Typed errors with context
- Error recovery strategies
- Retry mechanisms
- Component-specific error handling

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "your-project/agent"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()
    
    // Create configuration
    config := &agent.AgentConfig{
        ToolCallingModel: yourChatModel,
        ToolsConfig: map[string]tool.Tool{
            "calculator": yourCalculatorTool,
            "search": yourSearchTool,
        },
        MaxStep: 10,
        DebugMode: true,
    }
    
    // Create agent
    reactAgent, err := agent.NewAgent(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Prepare messages
    messages := []*schema.Message{
        {
            Role: schema.User,
            Content: "What is 25 * 4 + 10?",
        },
    }
    
    // Generate response
    response, err := reactAgent.Generate(ctx, messages)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", response.Content)
}
```

### Advanced Usage with Streaming

```go
// Stream responses
stream, err := reactAgent.Stream(ctx, messages)
if err != nil {
    log.Fatal(err)
}

for {
    chunk, err := stream.Recv()
    if err != nil {
        if err.Error() == "EOF" {
            break
        }
        log.Printf("Stream error: %v", err)
        break
    }
    fmt.Printf("Chunk: %s\n", chunk.Content)
}
```

### Monitoring and Metrics

```go
// Get execution metrics
if reactAgent.Monitor().IsEnabled() {
    metrics := reactAgent.Monitor().GetMetrics()
    fmt.Printf("Success Rate: %.2f%%\n", reactAgent.Monitor().GetSuccessRate())
    fmt.Printf("Average Execution Time: %v\n", metrics.AverageExecutionTime)
    fmt.Printf("Total Tool Calls: %d\n", metrics.ToolCallCount)
    
    // Print detailed summary
    reactAgent.Monitor().PrintSummary()
}
```

## Key Improvements

### 1. Proper Tool Integration
- Tools are now properly integrated into the execution flow
- Tool calls are handled through the decision controller
- Tool results are properly formatted and fed back into reasoning

### 2. Enhanced Graph Orchestration
- Based on official `eino` ReAct implementation
- Proper node-based execution with conditional routing
- Support for complex multi-step reasoning

### 3. Comprehensive Error Handling
- Structured error types with context
- Automatic retry mechanisms
- Component-specific error recovery

### 4. Advanced Monitoring
- Real-time execution tracking
- Performance metrics collection
- Debug logging with multiple levels

### 5. Eliminated Duplicates
- Single `ExecutionStep` definition
- Consolidated state management
- Unified error handling

## Configuration Options

### AgentConfig Fields

- `ToolCallingModel`: The chat model for reasoning
- `ToolsConfig`: Map of available tools
- `MaxStep`: Maximum execution steps
- `DebugMode`: Enable debug logging
- `MessageModifier`: Function to modify messages
- `StreamToolCallChecker`: Function for streaming support
- `GraphOptions`: Additional graph configuration

### Monitor Configuration

- `LogLevel`: Debug, Info, Warn, Error
- `Enabled`: Enable/disable monitoring
- `Metrics`: Automatic metrics collection

### Error Recovery Configuration

- `MaxRetries`: Maximum retry attempts
- `RetryDelay`: Base delay between retries
- `Recoverable`: Map of recoverable error types

## Best Practices

1. **Always enable monitoring** in development for debugging
2. **Use structured error handling** to catch and handle specific error types
3. **Configure appropriate retry policies** for production use
4. **Implement proper tool validation** to ensure reliable execution
5. **Use streaming** for long-running tasks to provide real-time feedback

## Migration from Old Structure

If migrating from the old multi-directory structure:

1. Update imports to use the single `agent` package
2. Replace duplicate type definitions with the consolidated ones
3. Update error handling to use the new structured errors
4. Enable monitoring for better observability
5. Configure retry policies for production resilience

## Testing

See `example.go` for comprehensive usage examples including:
- Basic agent usage
- Advanced multi-step tasks
- Error handling scenarios
- Monitoring and metrics collection

## Dependencies

- `github.com/cloudwego/eino`: Core framework for agent orchestration
- Standard Go libraries for logging, time, context management

This refactored implementation provides a robust, maintainable, and feature-rich ReAct agent that follows best practices and integrates seamlessly with the `eino` ecosystem.