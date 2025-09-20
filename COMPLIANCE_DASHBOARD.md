# Compliance Dashboard for Forge

A comprehensive compliance dashboard designed for compliance officers to monitor historical data and status reports in the Forge CI/CD platform.

## Overview

The Compliance Dashboard provides a centralized view of system compliance metrics, user activity, task execution status, and security events. It's designed specifically for compliance officers who need to monitor and report on system usage, security, and operational compliance.

## Features

### 📊 Dashboard Overview
- **Summary Cards**: High-level metrics including total tasks, success rates, active users, and security incidents
- **Interactive Charts**: Visual representation of trends over time with multiple chart types
- **Real-time Data**: Live updates of compliance metrics and status
- **Export Functionality**: Generate compliance reports in CSV format

### 🔍 Data Views

#### Task Compliance
- Task execution status and compliance scores
- Duration analysis and performance metrics
- User attribution and project association
- Issue tracking and resolution status

#### User Activity
- User engagement and activity patterns
- Role-based access monitoring
- External vs internal user tracking
- Last activity timestamps and frequency

#### Project Compliance
- Project-level success rates and metrics
- Team size and collaboration metrics
- Compliance scoring and issue identification
- Activity patterns and trends

#### Security Events
- Security incident tracking and categorization
- Severity levels and resolution status
- User and project attribution
- Event timeline and impact analysis

### 📈 Analytics & Reporting

#### Trend Analysis
- Daily task execution trends
- Success rate progression over time
- User activity patterns
- Security event frequency

#### Compliance Metrics
- Overall system compliance score
- Project-specific compliance rates
- User engagement metrics
- Security posture assessment

## Technical Implementation

### Backend API (`api/compliance.go`)

The compliance dashboard is powered by a comprehensive Go API that provides:

```go
// Main dashboard endpoint
GET /api/compliance/dashboard?days=30&project_id=123

// Export functionality
GET /api/compliance/dashboard/export?format=csv&days=30

// Individual data endpoints
GET /api/compliance/tasks/{taskId}
GET /api/compliance/users/{userId}
GET /api/compliance/projects/{projectId}
```

#### Data Models

**ComplianceDashboardData**
```go
type ComplianceDashboardData struct {
    Summary              ComplianceSummary
    TaskCompliance       []TaskComplianceData
    UserActivity         []UserActivityData
    ProjectCompliance    []ProjectComplianceData
    SecurityEvents       []SecurityEventData
    ComplianceTrends     ComplianceTrendsData
    LastUpdated          time.Time
}
```

**ComplianceSummary**
```go
type ComplianceSummary struct {
    TotalTasks           int
    SuccessfulTasks      int
    FailedTasks          int
    SuccessRate          float64
    TotalUsers           int
    ActiveUsers          int
    TotalProjects        int
    CompliantProjects    int
    ComplianceRate       float64
    SecurityIncidents    int
    LastAuditDate        *time.Time
}
```

### Frontend Components

#### Main Dashboard (`web/src/views/ComplianceDashboard.vue`)
- Central dashboard view with summary cards and charts
- Filter controls for date range and project selection
- Export functionality for compliance reports
- Responsive design for mobile and desktop

#### Chart Components (`web/src/components/compliance/`)
- **ComplianceTrendsChart.vue**: Interactive charts with multiple visualization types
- **TaskComplianceTable.vue**: Detailed task compliance data table
- **UserActivityTable.vue**: User activity and engagement metrics
- **ProjectComplianceTable.vue**: Project-level compliance analysis
- **SecurityEventsTable.vue**: Security incident tracking and management

#### Service Layer (`web/src/lib/complianceService.js`)
- Centralized API communication
- Data processing and transformation
- Error handling and user feedback
- Export functionality and file downloads

### Styling (`web/src/assets/scss/compliance-dashboard.scss`)

Comprehensive SCSS styling including:
- Responsive design for all screen sizes
- Dark mode support
- Animation and transition effects
- Accessibility considerations
- Modern gradient designs and card layouts

## Installation & Setup

### Prerequisites
- Go 1.24.2+
- Node.js 16+
- Vue.js 2.x
- Vuetify 2.x

### Backend Setup

1. **Add API Routes** (already implemented in `api/router.go`):
```go
adminAPI.Path("/compliance/dashboard").HandlerFunc(GetComplianceDashboard).Methods("GET", "HEAD")
```

2. **Import Compliance Module** (already implemented in `api/compliance.go`):
```go
import "github.com/Digital-Data-Co/forge/api"
```

### Frontend Setup

1. **Add Route** (already implemented in `web/src/router/index.js`):
```javascript
{
  path: '/compliance',
  component: ComplianceDashboard,
}
```

2. **Add Navigation Menu** (already implemented in `web/src/App.vue`):
```vue
<v-list-item key="compliance" to="/compliance" v-if="user.admin">
  <v-list-item-icon>
    <v-icon>mdi-shield-check</v-icon>
  </v-list-item-icon>
  <v-list-item-content>
    Compliance Dashboard
  </v-list-item-content>
</v-list-item>
```

3. **Import Styles** (already implemented):
```scss
@import '@/assets/scss/compliance-dashboard.scss';
```

## Usage

### Accessing the Dashboard

1. **Admin Access Required**: Only users with admin privileges can access the compliance dashboard
2. **Navigation**: Access via the main navigation menu under "Compliance Dashboard"
3. **URL**: Navigate to `/compliance` in your Forge instance

### Dashboard Features

#### Filtering and Date Ranges
- **Date Range Selection**: Choose from last week, month, quarter, or year
- **Project Filtering**: Filter data by specific projects or view all projects
- **Real-time Updates**: Refresh data to get the latest compliance metrics

#### Exporting Reports
- **CSV Export**: Download comprehensive compliance reports
- **Date Range Selection**: Export data for specific time periods
- **Project Filtering**: Include or exclude specific projects from exports

#### Chart Interactions
- **Hover Details**: Hover over chart elements for detailed information
- **Zoom and Pan**: Interactive chart controls for detailed analysis
- **Multiple Chart Types**: Line, area, bar, and pie chart visualizations

## Data Security & Privacy

### Access Control
- **Admin Only**: Compliance dashboard requires admin privileges
- **Project Scoping**: Data is filtered based on user permissions
- **Audit Logging**: All dashboard access is logged for compliance tracking

### Data Protection
- **Sensitive Data**: User passwords and secrets are never exposed
- **Anonymization**: Personal data is handled according to privacy requirements
- **Retention**: Data retention policies are applied to compliance reports

## Performance Considerations

### Backend Optimization
- **Database Indexing**: Optimized queries for large datasets
- **Caching**: Strategic caching of frequently accessed data
- **Pagination**: Large datasets are paginated for performance

### Frontend Optimization
- **Lazy Loading**: Charts and tables load on demand
- **Virtual Scrolling**: Large data tables use virtual scrolling
- **Debounced Updates**: Search and filter operations are debounced

## Customization

### Adding New Metrics
1. **Backend**: Add new fields to compliance data structures
2. **API**: Extend compliance endpoints with new data
3. **Frontend**: Add new components or extend existing ones
4. **Styling**: Customize appearance with SCSS variables

### Chart Customization
```vue
<ComplianceTrendsChart 
  :data="chartData"
  title="Custom Chart"
  color="#custom-color"
  chart-type="bar"
  :height="400"
  :show-legend="true"
/>
```

## Troubleshooting

### Common Issues

1. **No Data Displayed**
   - Check admin privileges
   - Verify date range selection
   - Ensure projects exist with activity

2. **Charts Not Rendering**
   - Check browser console for JavaScript errors
   - Verify ECharts library is loaded
   - Ensure data format is correct

3. **Export Issues**
   - Check file download permissions
   - Verify browser popup blockers
   - Ensure sufficient data for export

### Debug Mode
Enable debug mode by adding `?debug=true` to the dashboard URL for additional logging and error information.

## Contributing

### Development Guidelines
1. **Code Style**: Follow existing Vue.js and Go conventions
2. **Testing**: Add unit tests for new components and API endpoints
3. **Documentation**: Update this README for new features
4. **Accessibility**: Ensure WCAG 2.1 compliance for new UI components

### Adding New Compliance Metrics
1. **Backend**: Extend `ComplianceDashboardData` struct
2. **API**: Add new data retrieval logic
3. **Frontend**: Create new components or extend existing ones
4. **Styling**: Add appropriate styling for new elements

## License

This compliance dashboard is part of the Forge project and follows the same licensing terms.

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Review the Forge documentation
3. Create an issue in the Forge repository
4. Contact the development team for enterprise support

---

**Note**: This compliance dashboard is designed for compliance officers and administrators. Ensure proper access controls and data protection measures are in place before deploying in production environments.
