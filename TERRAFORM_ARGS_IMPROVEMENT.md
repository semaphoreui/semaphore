# Terraform Multi-Stage Arguments Support

## Overview

Enhanced the argument handling system to support stage-specific CLI arguments for Terraform tasks. This allows providing different arguments for different Terraform stages (init, plan, apply) which is essential for complex Terraform workflows.

## What Changed

### 1. LocalAppRunningArgs Structure (`db_lib/LocalApp.go`)

Added support for map-based arguments alongside the existing array format:

```go
type LocalAppRunningArgs struct {
    CliArgs         []string                // Legacy array format (backward compatible)
    CliArgsMap      map[string][]string     // New map format for stage-specific args
    EnvironmentVars []string
    Inputs          map[string]string
    TaskParams      any
    TemplateParams  any
    Callback        func(*os.Process)
}
```

### 2. Argument Parsing (`services/tasks/LocalJob.go`)

Added `getCLIArgsMap()` function that:
- Attempts to parse arguments as an array first (backward compatible)
- Falls back to map format if array parsing fails
- Returns both array and map representations
- Supports both Template and Task level arguments

### 3. Terraform Argument Processing (`services/tasks/LocalJob.go`)

Updated `getTerraformArgs()` to:
- Return both array and map formats
- Merge template and task arguments at the stage level
- Apply common args (destroy, vars, secrets) to all stages
- Auto-create "init" stage if not explicitly defined

### 4. TerraformApp Enhancements (`db_lib/TerraformApp.go`)

Modified Terraform execution to:
- Accept stage-specific init args during installation
- Use different args for plan and apply stages
- Maintain backward compatibility with array format
- New method `InstallRequirementsWithInitArgs()` for init customization

### 5. LocalJob Orchestration (`services/tasks/LocalJob.go`)

Enhanced `Run()` method to:
- Get args before prepareRun for Terraform apps
- Pass init-specific args during installation
- Provide plan/apply-specific args during execution

## Usage Examples

### Legacy Format (Still Supported)

Array format arguments apply to all stages:

```json
{
  "arguments": ["-var", "environment=production"]
}
```

### New Map Format

Stage-specific arguments for different Terraform operations:

```json
{
  "arguments": {
    "init": ["-upgrade"],
    "plan": ["-var", "foo=bar"],
    "apply": ["-var", "foo=baz"]
  }
}
```

### Real-World Example

Template with stage-specific configurations:

```json
{
  "template": {
    "arguments": {
      "init": ["-backend-config=bucket=my-bucket"],
      "plan": ["-out=tfplan"],
      "apply": ["tfplan"]
    }
  }
}
```

Task override combining with template args:

```json
{
  "task": {
    "arguments": {
      "init": ["-reconfigure"],
      "apply": ["-auto-approve"]
    }
  }
}
```

Result: Arguments are merged per stage
- **init**: `-backend-config=bucket=my-bucket`, `-reconfigure`
- **plan**: `-out=tfplan`
- **apply**: `tfplan`, `-auto-approve`

## Backward Compatibility

✅ **100% Backward Compatible**

- Existing array format continues to work
- No changes required to existing templates/tasks
- Array format arguments are used for all stages when no map is provided
- Gradual migration path available

## Implementation Details

### Stage-Specific Argument Flow

1. **Parse Phase**: Arguments parsed as array or map from JSON
2. **Merge Phase**: Template and task args merged at stage level
3. **Common Args**: Environment vars, secrets, and destroy flag added to all stages
4. **Execution Phase**: Appropriate args used for each stage (init, plan, apply)

### Key Functions

- `getCLIArgsMap()`: Parses both formats from JSON
- `getTerraformArgs()`: Builds stage-specific argument maps
- `prepareRunTerraform()`: Passes init args to Terraform installation
- `TerraformApp.Run()`: Uses plan/apply-specific args during execution

### Supported Stages

- **init**: Used during `terraform init` (via InstallRequirements)
- **plan**: Used during `terraform plan`
- **apply**: Used during `terraform apply`

Note: If using map format, unspecified stages fall back to CliArgs if available.

## Testing

The implementation has been validated with:
- Successful build of entire project
- No linter errors
- Backward compatibility verified
- Both array and map formats tested

## Benefits

1. **Flexibility**: Different args for different Terraform stages
2. **Security**: Keep sensitive args only in specific stages
3. **Efficiency**: Optimize each stage independently
4. **Clarity**: Clear separation of stage-specific configurations
5. **Compatibility**: Works alongside existing array format

## Migration Path

### Phase 1: Keep using array format (no changes needed)
```json
{"arguments": ["-var", "foo=bar"]}
```

### Phase 2: Migrate to map format for multi-stage tasks
```json
{"arguments": {"init": ["-upgrade"], "apply": ["-var", "foo=bar"]}}
```

### Phase 3: Leverage full stage-specific capabilities
```json
{
  "arguments": {
    "init": ["-backend-config=..."],
    "plan": ["-out=tfplan", "-var-file=prod.tfvars"],
    "apply": ["tfplan", "-parallelism=20"]
  }
}
```

