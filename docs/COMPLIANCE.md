# OpenSCAP Compliance Integration

This document describes the OpenSCAP compliance integration in Forge, which provides
centralized policy management, periodic audits, ARF report collection, and compliance visualization.

## Overview

The OpenSCAP integration allows you to:
- Upload SCAP DataStream files and discover available profiles
- Create compliance policies targeting specific hosts or inventories
- Schedule periodic compliance scans
- View detailed compliance reports and rule results
- Download ARF (Asset Reporting Format) files for external analysis

## Prerequisites

### OpenSCAP Installation

OpenSCAP must be installed on all runner hosts where compliance scans will be executed.

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install openscap-scanner scap-security-guide
```

#### CentOS/RHEL/Rocky Linux
```bash
sudo yum install openscap-scanner scap-security-guide
# or for newer versions:
sudo dnf install openscap-scanner scap-security-guide
```

#### Fedora
```bash
sudo dnf install openscap-scanner scap-security-guide
```

#### SLES/openSUSE
```bash
sudo zypper install openscap-scanner scap-security-guide
```

### Verification

To verify OpenSCAP installation, run:
```bash
oscap --version
oscap info --help
```

## Usage

### 1. Upload SCAP Content

1. Navigate to **Compliance > Contents** in your project
2. Click **Upload Content**
3. Select a SCAP DataStream (.xml) file
4. Provide a name and optional source description
5. Click **Upload**

The system will automatically:
- Validate the SCAP content
- Extract available profiles
- Store the content securely

### 2. Create Compliance Policies

1. Navigate to **Compliance > Policies**
2. Click **Create Policy**
3. Fill in the policy details:
   - **Name**: Descriptive name for the policy
   - **Content**: Select uploaded SCAP content
   - **Profile**: Choose a profile from the content
   - **Schedule**: Optional cron schedule for periodic scans
4. Click **Create**

### 3. Assign Policies to Targets

1. Select a policy from the policies list
2. Click **Assignments**
3. Add targets:
   - **Inventory**: Apply to all hosts in an inventory
   - **Host**: Apply to specific hosts
   - **Group**: Apply to host groups
4. Save assignments

### 4. Run Compliance Scans

#### Manual Scans
1. Select a policy
2. Click **Scan Now**

#### Scheduled Scans
Policies with schedules will automatically run at the specified intervals.

### 5. View Compliance Reports

1. Navigate to **Compliance > Reports**
2. Filter by:
   - Policy
   - Host
   - Result status
   - Date range
3. Click on a report to view details
4. Download ARF files for external analysis

## API Endpoints

### SCAP Content
- `GET /api/project/{project_id}/compliance/contents` - List contents
- `POST /api/project/{project_id}/compliance/contents` - Upload content
- `GET /api/project/{project_id}/compliance/contents/{id}` - Get content details
- `GET /api/project/{project_id}/compliance/contents/{id}/profiles` - Get profiles
- `DELETE /api/project/{project_id}/compliance/contents/{id}` - Delete content

### Compliance Policies
- `GET /api/project/{project_id}/compliance/policies` - List policies
- `POST /api/project/{project_id}/compliance/policies` - Create policy
- `GET /api/project/{project_id}/compliance/policies/{id}` - Get policy details
- `PUT /api/project/{project_id}/compliance/policies/{id}` - Update policy
- `DELETE /api/project/{project_id}/compliance/policies/{id}` - Delete policy
- `GET /api/project/{project_id}/compliance/policies/{id}/assignments` - Get assignments
- `PUT /api/project/{project_id}/compliance/policies/{id}/assignments` - Set assignments
- `POST /api/project/{project_id}/compliance/policies/{id}/scan` - Run scan

### Compliance Scans
- `GET /api/project/{project_id}/compliance/scans` - List scans
- `GET /api/project/{project_id}/compliance/scans/{id}` - Get scan details
- `POST /api/project/{project_id}/compliance/scans/{id}` - Cancel scan
- `GET /api/project/{project_id}/compliance/scans/{id}/reports` - Get scan reports

### Compliance Reports
- `GET /api/project/{project_id}/compliance/reports` - List reports
- `GET /api/project/{project_id}/compliance/reports/{id}` - Get report details
- `GET /api/project/{project_id}/compliance/reports/{id}/arf` - Download ARF file

### System Health
- `GET /api/project/{project_id}/compliance/preflight` - Check OpenSCAP installation

## Configuration

### File Storage

SCAP content files are stored in:
```
uploads/scap/project_{project_id}/
```

Scan results are stored in:
```
scans/scan_{scan_id}/
```

### Database Tables

The following tables are created for compliance data:
- `scap_contents` - SCAP DataStream files
- `scap_profiles` - Discovered profiles
- `compliance_policies` - Compliance policies
- `policy_assignments` - Policy target assignments
- `compliance_scans` - Scan executions
- `compliance_reports` - Individual host results
- `compliance_rule_results` - Detailed rule results

## Troubleshooting

### OpenSCAP Not Found
If you see "oscap not found" errors:
1. Verify OpenSCAP is installed: `which oscap`
2. Check the PATH environment variable
3. Ensure OpenSCAP is accessible to the runner user

### Scan Failures
Common scan failure reasons:
1. **Permission denied**: Ensure runner has access to target hosts
2. **Network issues**: Check connectivity to target hosts
3. **Invalid content**: Verify SCAP DataStream is valid
4. **Missing dependencies**: Ensure required SCAP content is available

### Performance Issues
For large environments:
1. Limit concurrent scans per runner
2. Use appropriate scan schedules
3. Consider host grouping for better organization
4. Monitor disk usage for scan results

## Security Considerations

1. **Content Validation**: All uploaded SCAP content is validated before storage
2. **Access Control**: Compliance features respect project-level permissions
3. **File Storage**: Content files are stored with restricted permissions
4. **Network Security**: Ensure secure communication with target hosts

## Integration with Existing Features

The compliance system integrates with:
- **Runners**: Uses existing runner infrastructure for scan execution
- **Inventories**: Can target existing inventory hosts
- **Scheduling**: Uses the existing schedule system for periodic scans
- **Tasks**: Creates tasks for scan tracking and logging
- **File Management**: Uses the existing task file system for ARF storage

## Limitations

1. **Remote Execution**: Currently supports local execution only
   (remote execution via SSH would require additional implementation)
2. **Content Types**: Supports XCCDF DataStream format only
3. **Profile Discovery**: Automatic profile discovery requires valid SCAP content
4. **Rule Results**: Detailed rule results depend on ARF parsing accuracy

## Future Enhancements

Planned improvements include:
- Remote execution via SSH
- Additional SCAP content format support
- Enhanced reporting and visualization
- Integration with external compliance tools
- Automated remediation suggestions
