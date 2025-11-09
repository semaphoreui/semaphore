# Terraform Multi-Stage Arguments Support

## Overview

Enhanced the argument handling system to support stage-specific CLI arguments for Terraform tasks. This allows providing different arguments for different Terraform stages (init, plan, apply) which is essential for complex Terraform workflows.

## What Changed

### 1. LocalAppRunningArgs Structure (`db_lib/LocalApp.go`)

Unified to use map-based arguments with "default" key for backward compatibility:

```go
type LocalAppRunningArgs struct {
    CliArgs         map[string][]string     // Stage-specific args (e.g., "init", "apply", "default")
    EnvironmentVars []string
    Inputs          map[string]string
    TaskParams      any
    TemplateParams  any
    Callback        func(*os.Process)
}
```

**Key Change**: Array format arguments are automatically converted to map format with key "default".

### 2. Argument Parsing (`services/tasks/LocalJob.go`)

Added `convertArgsJSONIfArray()` and `getCLIArgsMap()` functions that:
- `convertArgsJSONIfArray()`: Checks JSON format and converts array format to map with "default" key **in-place**
- `getCLIArgsMap()`: Parses arguments as map format (after conversion)
- Array format is automatically converted to map format at runtime
- Supports both Template and Task level arguments
- Ensures consistent map-based interface throughout the system

### 3. Terraform Argument Processing (`services/tasks/LocalJob.go`)

Updated `getTerraformArgs()` to:
- Return map format only (unified interface)
- Merge template and task arguments at the stage level
- Apply common args (destroy, vars, secrets) to all stages
- Ensure at least "default" stage exists with common args

### 4. TerraformApp Enhancements (`db_lib/TerraformApp.go`)

Modified Terraform execution to:
- Accept stage-specific init args during installation
- Use different args for plan and apply stages
- Fall back to "default" key when specific stage not defined
- New method `InstallRequirementsWithInitArgs()` for init customization

### 5. LocalJob Orchestration (`services/tasks/LocalJob.go`)

Enhanced `Run()` method to:
- Get args before prepareRun for Terraform apps
- Pass init-specific args during installation
- Provide plan/apply-specific args during execution
- Convert all args to unified map format with "default" key for non-Terraform apps

## Usage Examples

### Legacy Format (Still Supported)

Array format arguments are automatically converted to map with "default" key:

```json
{
  "arguments": ["-var", "environment=production"]
}
```

**Internally converted to:**
```json
{
  "arguments": {
    "default": ["-var", "environment=production"]
  }
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
- **default**: Used as fallback when specific stage not defined

### Stage Resolution Order (Terraform)

For each stage, arguments are resolved in this order:
1. Stage-specific key (e.g., "init", "plan", "apply")
2. Fall back to "default" key if stage-specific not found
3. Empty array if neither exists

### Backward Compatibility Details

**Array Format → Map Conversion:**
- **Runtime Conversion**: Array `["-var", "foo=bar"]` is converted to `{"default": ["-var", "foo=bar"]}` **in-place** when the task runs
- **No Database Changes**: Original JSON remains stored as array, conversion happens only during task execution
- **Transparent**: Users don't see the conversion, it happens automatically
- Ansible and Shell apps: Always use "default" key
- Terraform apps: Use stage-specific keys, fall back to "default"

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
**Result:** Automatically converted to `{"default": ["-var", "foo=bar"]}` internally

### Phase 2: Migrate to map format for multi-stage tasks
```json
{"arguments": {"init": ["-upgrade"], "apply": ["-var", "foo=bar"]}}
```
**Result:** Uses stage-specific args for init and apply

### Phase 3: Mix default and stage-specific for flexibility
```json
{
  "arguments": {
    "default": ["-var", "common=value"],
    "init": ["-upgrade"],
    "apply": ["-parallelism=20"]
  }
}
```
**Result:** plan stage uses "default" args, init and apply use their specific args

### Phase 4: Leverage full stage-specific capabilities
```json
{
  "arguments": {
    "init": ["-backend-config=..."],
    "plan": ["-out=tfplan", "-var-file=prod.tfvars"],
    "apply": ["tfplan", "-parallelism=20"]
  }
}
```
**Result:** Complete control over each stage independently

