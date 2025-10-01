/**
 * Terraform Generator Utility
 * Converts JSON output from ModuleSelector into valid Terraform configurations
 */

// Cloud provider specific configurations
const getCloudProviderModule = {
  AWS: 'cloud-aws',
  Azure: 'cloud-azure',
  'Google Cloud': 'cloud-gcp',
};

// Terraform resource templates
const terraformTemplates = {
  aws: {
    main: () => `
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "main" {
  source = "../../modules/cloud-aws"

  region             = var.aws_region
  tags               = var.tags
  create_example     = var.create_example
  example_group_name = var.example_group_name

  # Kubernetes configuration
  kubernetes_type      = var.kubernetes_type
  instance_type        = var.instance_type
  control_plane_nodes  = var.control_plane_nodes
  min_workers         = var.min_workers
  max_workers         = var.max_workers
  golden_image        = var.golden_image
  stig_compliant      = var.stig_compliant
  jumpbox             = var.jumpbox
  jumpbox_stig_compliant = var.jumpbox_stig_compliant

  # Additional software
  enable_observability     = var.enable_observability
  enable_service_mesh      = var.enable_service_mesh
  enable_certificate_manager = var.enable_certificate_manager
  enable_gateway_api       = var.enable_gateway_api
  enable_nginx_ingress_proxy = var.enable_nginx_ingress_proxy
}`,
    variables: () => `
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

variable "create_example" {
  description = "Create example resources"
  type        = bool
  default     = true
}

variable "example_group_name" {
  description = "Name for example group"
  type        = string
  default     = "tf-module-example"
}

variable "kubernetes_type" {
  description = "Kubernetes type"
  type        = string
  default     = "EKS"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "min_workers" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 3
}

variable "max_workers" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 6
}

variable "golden_image" {
  description = "Golden image to use"
  type        = string
  default     = "Amazon Linux 2"
}

variable "stig_compliant" {
  description = "Enable STIG compliance"
  type        = bool
  default     = false
}

variable "control_plane_nodes" {
  description = "Number of control plane nodes (for self-managed Kubernetes)"
  type        = number
  default     = 3
}

variable "jumpbox" {
  description = "Jumpbox operating system"
  type        = string
  default     = "Windows 11 Pro"
}

variable "jumpbox_stig_compliant" {
  description = "Enable STIG compliance for jumpbox"
  type        = bool
  default     = true
}

variable "enable_observability" {
  description = "Enable observability stack"
  type        = bool
  default     = false
}

variable "enable_service_mesh" {
  description = "Enable service mesh"
  type        = bool
  default     = false
}

variable "enable_certificate_manager" {
  description = "Enable certificate manager"
  type        = bool
  default     = false
}

variable "enable_gateway_api" {
  description = "Enable Gateway API"
  type        = bool
  default     = false
}

variable "enable_nginx_ingress_proxy" {
  description = "Enable Nginx Ingress Proxy"
  type        = bool
  default     = false
}`,
    outputs: () => `
output "cluster_id" {
  description = "EKS cluster ID"
  value       = module.main.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.main.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ids attached to the cluster control plane"
  value       = module.main.cluster_security_group_id
}
`,
  },

  azure: {
    main: () => `
terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
  location = var.location
}

module "main" {
  source = "../../modules/cloud-azure"

  location = var.location
  tags     = var.tags

  # Kubernetes configuration
  kubernetes_type      = var.kubernetes_type
  vm_size              = var.vm_size
  control_plane_nodes  = var.control_plane_nodes
  min_workers         = var.min_workers
  max_workers         = var.max_workers
  golden_image        = var.golden_image
  stig_compliant      = var.stig_compliant
  jumpbox             = var.jumpbox
  jumpbox_stig_compliant = var.jumpbox_stig_compliant

  # Additional software
  enable_observability     = var.enable_observability
  enable_service_mesh      = var.enable_service_mesh
  enable_certificate_manager = var.enable_certificate_manager
  enable_gateway_api       = var.enable_gateway_api
  enable_nginx_ingress_proxy = var.enable_nginx_ingress_proxy
}`,
    variables: () => `
variable "location" {
  description = "Azure region"
  type        = string
  default     = "East US"
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

variable "kubernetes_type" {
  description = "Kubernetes type (AKS for managed)"
  type        = string
  default     = "AKS"
}

variable "vm_size" {
  description = "VM size for worker nodes"
  type        = string
  default     = "Standard_D2s_v3"
}

variable "min_workers" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 3
}

variable "max_workers" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 6
}

variable "golden_image" {
  description = "Golden image to use"
  type        = string
  default     = "Ubuntu"
}

variable "stig_compliant" {
  description = "Enable STIG compliance"
  type        = bool
  default     = false
}

variable "control_plane_nodes" {
  description = "Number of control plane nodes (for self-managed Kubernetes)"
  type        = number
  default     = 3
}

variable "jumpbox" {
  description = "Jumpbox operating system"
  type        = string
  default     = "Windows 11 Pro"
}

variable "jumpbox_stig_compliant" {
  description = "Enable STIG compliance for jumpbox"
  type        = bool
  default     = true
}

variable "enable_observability" {
  description = "Enable observability stack"
  type        = bool
  default     = false
}

variable "enable_service_mesh" {
  description = "Enable service mesh"
  type        = bool
  default     = false
}

variable "enable_certificate_manager" {
  description = "Enable certificate manager"
  type        = bool
  default     = false
}

variable "enable_gateway_api" {
  description = "Enable Gateway API"
  type        = bool
  default     = false
}

variable "enable_nginx_ingress_proxy" {
  description = "Enable Nginx Ingress Proxy"
  type        = bool
  default     = false
}`,
    outputs: () => `
output "cluster_id" {
  description = "AKS cluster ID"
  value       = module.main.cluster_id
}

output "cluster_fqdn" {
  description = "FQDN of the AKS cluster"
  value       = module.main.cluster_fqdn
}

output "cluster_private_fqdn" {
  description = "FQDN of the AKS cluster when private cluster is enabled"
  value       = module.main.cluster_private_fqdn
}
`,
  },

  gcp: {
    main: () => `
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

module "main" {
  source = "../../modules/cloud-gcp"

  project_id = var.project_id
  region     = var.region
  zone       = var.zone
  tags       = var.tags

  # Kubernetes configuration
  kubernetes_type      = var.kubernetes_type
  machine_type         = var.machine_type
  control_plane_nodes  = var.control_plane_nodes
  min_node_count      = var.min_node_count
  max_node_count      = var.max_node_count
  golden_image        = var.golden_image
  stig_compliant      = var.stig_compliant
  jumpbox             = var.jumpbox
  jumpbox_stig_compliant = var.jumpbox_stig_compliant

  # Additional software
  enable_observability     = var.enable_observability
  enable_service_mesh      = var.enable_service_mesh
  enable_certificate_manager = var.enable_certificate_manager
  enable_gateway_api       = var.enable_gateway_api
  enable_nginx_ingress_proxy = var.enable_nginx_ingress_proxy
}`,
    variables: () => `
variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default = {
    environment = "production"
    managed-by  = "terraform"
  }
}

variable "kubernetes_type" {
  description = "Kubernetes type (GKE for managed)"
  type        = string
  default     = "GKE"
}

variable "machine_type" {
  description = "Machine type for worker nodes"
  type        = string
  default     = "e2-standard-2"
}

variable "min_node_count" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 3
}

variable "max_node_count" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 6
}

variable "golden_image" {
  description = "Golden image to use"
  type        = string
  default     = "Ubuntu"
}

variable "stig_compliant" {
  description = "Enable STIG compliance"
  type        = bool
  default     = false
}

variable "control_plane_nodes" {
  description = "Number of control plane nodes (for self-managed Kubernetes)"
  type        = number
  default     = 3
}

variable "jumpbox" {
  description = "Jumpbox operating system"
  type        = string
  default     = "Windows 11 Pro"
}

variable "jumpbox_stig_compliant" {
  description = "Enable STIG compliance for jumpbox"
  type        = bool
  default     = true
}

variable "enable_observability" {
  description = "Enable observability stack"
  type        = bool
  default     = false
}

variable "enable_service_mesh" {
  description = "Enable service mesh"
  type        = bool
  default     = false
}

variable "enable_certificate_manager" {
  description = "Enable certificate manager"
  type        = bool
  default     = false
}

variable "enable_gateway_api" {
  description = "Enable Gateway API"
  type        = bool
  default     = false
}

variable "enable_nginx_ingress_proxy" {
  description = "Enable Nginx Ingress Proxy"
  type        = bool
  default     = false
}`,
    outputs: () => `
output "cluster_id" {
  description = "GKE cluster ID"
  value       = module.main.cluster_id
}

output "cluster_endpoint" {
  description = "GKE cluster endpoint"
  value       = module.main.cluster_endpoint
}

output "cluster_ca_certificate" {
  description = "Base64 encoded cluster certificate"
  value       = module.main.cluster_ca_certificate
}
`,
  },
};

// Generate terraform.tfvars.example file
function generateTfvarsExample(config, provider) {
  const baseConfig = `# Example Terraform variables file
# Copy to terraform.tfvars and update values

cloud_provider = "${config.cloudProvider}"
kubernetes_type = "${config.kubernetesType}"`;

  switch (provider) {
    case 'aws':
      return `${baseConfig}

aws_region = "us-east-1"
instance_type = "${config.instanceType}"
control_plane_nodes = ${config.controlPlaneNodes}
min_workers = ${config.minWorkers}
max_workers = ${config.maxWorkers}
golden_image = "${config.goldenImage}"
stig_compliant = ${config.isStigCompliant}
jumpbox = "${config.jumpbox}"
jumpbox_stig_compliant = ${config.jumpboxStigCompliant}

# Additional software
enable_observability = ${config.additionalSoftware.observability}
enable_service_mesh = ${config.additionalSoftware.serviceMesh}
enable_certificate_manager = ${config.additionalSoftware.certificateManager}
enable_gateway_api = ${config.additionalSoftware.gatewayApi}
enable_nginx_ingress_proxy = ${config.additionalSoftware.nginxIngressProxy}
create_example = true
example_group_name = "dd-example-group"
tags = {
  Environment = "production"
  ManagedBy   = "terraform"
}`;

    case 'azure':
      return `${baseConfig}

location = "East US"
vm_size = "${config.instanceType}"
control_plane_nodes = ${config.controlPlaneNodes}
min_workers = ${config.minWorkers}
max_workers = ${config.maxWorkers}
golden_image = "${config.goldenImage}"
stig_compliant = ${config.isStigCompliant}
jumpbox = "${config.jumpbox}"
jumpbox_stig_compliant = ${config.jumpboxStigCompliant}

# Additional software
enable_observability = ${config.additionalSoftware.observability}
enable_service_mesh = ${config.additionalSoftware.serviceMesh}
enable_certificate_manager = ${config.additionalSoftware.certificateManager}
enable_gateway_api = ${config.additionalSoftware.gatewayApi}
enable_nginx_ingress_proxy = ${config.additionalSoftware.nginxIngressProxy}
tags = {
  Environment = "production"
  ManagedBy   = "terraform"
}`;

    case 'gcp':
      return `${baseConfig}

project_id = "your-gcp-project-id"
region = "us-central1"
zone = "us-central1-a"
machine_type = "${config.instanceType}"
control_plane_nodes = ${config.controlPlaneNodes}
min_node_count = ${config.minWorkers}
max_node_count = ${config.maxWorkers}
golden_image = "${config.goldenImage}"
stig_compliant = ${config.isStigCompliant}
jumpbox = "${config.jumpbox}"
jumpbox_stig_compliant = ${config.jumpboxStigCompliant}

# Additional software
enable_observability = ${config.additionalSoftware.observability}
enable_service_mesh = ${config.additionalSoftware.serviceMesh}
enable_certificate_manager = ${config.additionalSoftware.certificateManager}
enable_gateway_api = ${config.additionalSoftware.gatewayApi}
enable_nginx_ingress_proxy = ${config.additionalSoftware.nginxIngressProxy}
tags = {
  environment = "production"
  managed-by  = "terraform"
}`;

    default:
      return baseConfig;
  }
}

// Generate Terragrunt inputs
function generateTerragruntInputs(config) {
  const {
    cloudProvider,
    instanceType,
    controlPlaneNodes,
    minWorkers,
    maxWorkers,
    goldenImage,
    isStigCompliant,
    jumpbox,
    jumpboxStigCompliant,
    additionalSoftware,
  } = config;

  const commonInputs = `
  kubernetes_type      = "${config.kubernetesType}"
  control_plane_nodes  = ${controlPlaneNodes}
  min_workers         = ${minWorkers}
  max_workers         = ${maxWorkers}
  golden_image        = "${goldenImage}"
  stig_compliant      = ${isStigCompliant}
  jumpbox             = "${jumpbox}"
  jumpbox_stig_compliant = ${jumpboxStigCompliant}
  enable_observability = ${additionalSoftware.observability}
  enable_service_mesh = ${additionalSoftware.serviceMesh}
  enable_certificate_manager = ${additionalSoftware.certificateManager}
  enable_gateway_api = ${additionalSoftware.gatewayApi}
  enable_nginx_ingress_proxy = ${additionalSoftware.nginxIngressProxy}`;

  switch (cloudProvider) {
    case 'AWS':
      return `${commonInputs}
  aws_region    = "us-east-1"
  instance_type  = "${instanceType}"
  create_example = true
  example_group_name = "dd-example-group"
  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }`;

    case 'Azure':
      return `${commonInputs}
  location = "East US"
  vm_size = "${instanceType}"
  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }`;

    case 'Google Cloud':
      return `${commonInputs}
  project_id    = "your-gcp-project-id"
  region        = "us-central1"
  machine_type  = "${instanceType}"
  tags = {
    environment = "production"
    managed-by  = "terraform"
  }`;

    default:
      return commonInputs;
  }
}

// Generate backend configuration for Terragrunt
function generateBackendHclExample(cloudProvider) {
  switch (cloudProvider) {
    case 'AWS':
      return `
# AWS S3 backend configuration
remote_state {
  backend = "s3"
  config = {
    bucket         = "your-terraform-state-bucket"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-locking"
  }
}`;

    case 'Azure':
      return `
# Azure backend configuration
remote_state {
  backend = "azurerm"
  config = {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "yourterraformstate"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}`;

    case 'Google Cloud':
      return `
# Google Cloud backend configuration
remote_state {
  backend = "gcs"
  config = {
    bucket = "your-terraform-state-bucket"
    prefix = "terraform/state"
  }
}`;

    default:
      return `
# Default local backend
remote_state {
  backend = "local"
  config = {
    path = "terraform.tfstate"
  }
}`;
  }
}

// Convert JSON configuration to Terraform
export function generateTerraformConfig(jsonOutput) {
  if (!jsonOutput) {
    throw new Error('JSON output is required to generate Terraform configuration');
  }

  const {
    cloudProvider,
    kubernetesType,
    instanceType,
    controlPlaneNodes,
    minWorkers,
    maxWorkers,
    goldenImage,
    isStigCompliant,
    jumpbox,
    jumpboxStigCompliant,
    additionalSoftware = {},
  } = jsonOutput;

  let providerKey = 'aws';
  switch (cloudProvider) {
    case 'AWS':
      providerKey = 'aws';
      break;
    case 'Azure':
      providerKey = 'azure';
      break;
    case 'Google Cloud':
      providerKey = 'gcp';
      break;
    default:
      throw new Error(`Unsupported cloud provider: ${cloudProvider}`);
  }

  const template = terraformTemplates[providerKey];
  if (!template) {
    throw new Error(`No Terraform template found for provider: ${providerKey}`);
  }

  // Create configuration object with only selected options
  const config = {
    cloudProvider,
    kubernetesType,
    instanceType,
    minWorkers,
    maxWorkers,
    goldenImage,
    isStigCompliant,
    jumpbox,
    jumpboxStigCompliant,
  };

  // Only include control plane nodes if Self-Managed Kubernetes is selected
  if (kubernetesType === 'Self-Managed Kubernetes' && controlPlaneNodes) {
    config.controlPlaneNodes = controlPlaneNodes;
  }

  // Only include additional software if any are selected
  if (additionalSoftware && Object.keys(additionalSoftware).length > 0) {
    config.additionalSoftware = additionalSoftware;
  }

  // Generate Terraform files
  const terraformFiles = {
    'main.tf': template.main().trim(),
    'variables.tf': template.variables().trim(),
    'outputs.tf': template.outputs().trim(),
    'tfvars.example': generateTfvarsExample(config, providerKey),
  };

  return terraformFiles;
}

// Generate HCL format (for Terragrunt)
export function generateTerragruntConfig(jsonOutput) {
  if (!jsonOutput) {
    throw new Error('JSON output is required to generate Terragrunt configuration');
  }

  const terraformFiles = generateTerraformConfig(jsonOutput);

  // Terragrunt specific configuration
  const providerKey = getCloudProviderModule[jsonOutput.cloudProvider] || 'aws';
  const terragruntHcl = `
terraform {
  source = "git::https://github.com/your-org/terraform-modules.git//modules/providers/${
  providerKey}?ref=main"
}

# Input variables for Terragrunt
inputs = {
  ${generateTerragruntInputs(jsonOutput)}
}
`.trim();

  return {
    terraform: terraformFiles,
    terragrunt: {
      'terragrunt.hcl': terragruntHcl,
      'backend.hcl.example': generateBackendHclExample(jsonOutput.cloudProvider),
    },
  };
}

export default {
  generateTerraformConfig,
  generateTerragruntConfig,
};
