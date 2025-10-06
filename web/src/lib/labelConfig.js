// Label configuration for hierarchical organization and color coding
export const LABEL_CATEGORIES = {
  priority: {
    name: 'Priority',
    color: '#FF5722',
    icon: 'mdi-flag',
    labels: [
      { name: 'critical', color: '#F44336', description: 'Critical priority tasks' },
      { name: 'high', color: '#FF9800', description: 'High priority tasks' },
      { name: 'medium', color: '#FFC107', description: 'Medium priority tasks' },
      { name: 'low', color: '#4CAF50', description: 'Low priority tasks' },
    ],
  },
  environment: {
    name: 'Environment',
    color: '#2196F3',
    icon: 'mdi-server',
    labels: [
      { name: 'production', color: '#F44336', description: 'Production environment' },
      { name: 'staging', color: '#FF9800', description: 'Staging environment' },
      { name: 'development', color: '#4CAF50', description: 'Development environment' },
      { name: 'testing', color: '#9C27B0', description: 'Testing environment' },
    ],
  },
  type: {
    name: 'Type',
    color: '#9C27B0',
    icon: 'mdi-tag',
    labels: [
      { name: 'infrastructure', color: '#607D8B', description: 'Infrastructure related' },
      { name: 'security', color: '#F44336', description: 'Security related' },
      { name: 'monitoring', color: '#FF9800', description: 'Monitoring related' },
      { name: 'deployment', color: '#4CAF50', description: 'Deployment related' },
      { name: 'backup', color: '#9C27B0', description: 'Backup related' },
      { name: 'maintenance', color: '#795548', description: 'Maintenance related' },
    ],
  },
  status: {
    name: 'Status',
    color: '#00BCD4',
    icon: 'mdi-information',
    labels: [
      { name: 'active', color: '#4CAF50', description: 'Active tasks' },
      { name: 'deprecated', color: '#9E9E9E', description: 'Deprecated tasks' },
      { name: 'experimental', color: '#FF9800', description: 'Experimental tasks' },
      { name: 'stable', color: '#2196F3', description: 'Stable tasks' },
    ],
  },
};

// Get label configuration by name
export function getLabelConfig(labelName) {
  const categories = Object.values(LABEL_CATEGORIES);
  for (let i = 0; i < categories.length; i += 1) {
    const category = categories[i];
    const label = category.labels.find((l) => l.name === labelName);
    if (label) {
      return {
        ...label,
        category: category.name,
        categoryColor: category.color,
        categoryIcon: category.icon,
      };
    }
  }
  return {
    name: labelName,
    color: '#9E9E9E',
    description: 'Custom label',
    category: 'Custom',
    categoryColor: '#9E9E9E',
    categoryIcon: 'mdi-tag',
  };
}

// Get all available labels
export function getAllLabels() {
  const allLabels = [];
  Object.values(LABEL_CATEGORIES).forEach((category) => {
    allLabels.push(...category.labels.map((label) => ({
      ...label,
      category: category.name,
      categoryColor: category.color,
      categoryIcon: category.icon,
    })));
  });
  return allLabels;
}

// Get labels by category
export function getLabelsByCategory(categoryName) {
  const category = LABEL_CATEGORIES[categoryName];
  if (!category) return [];

  return category.labels.map((label) => ({
    ...label,
    category: category.name,
    categoryColor: category.color,
    categoryIcon: category.icon,
  }));
}

// Validate label format (category:label)
export function parseLabel(labelString) {
  if (labelString.includes(':')) {
    const [category, label] = labelString.split(':', 2);
    return { category: category.trim(), label: label.trim() };
  }
  return { category: null, label: labelString.trim() };
}

// Format label for display
export function formatLabel(category, label) {
  if (category && label) {
    return `${category}:${label}`;
  }
  return label || category;
}
