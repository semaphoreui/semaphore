import axios from 'axios';

/**
 * Compliance Dashboard Service
 * Handles all compliance-related API calls and data processing
 */
class ComplianceService {
  constructor() {
    this.baseURL = '/api/compliance';
  }

  /**
   * Get compliance dashboard data
   * @param {Object} params - Query parameters
   * @param {number} params.days - Number of days to look back
   * @param {string} params.project_id - Optional project ID filter
   * @returns {Promise<Object>} Compliance dashboard data
   */
  async getDashboardData(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/dashboard`, { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch compliance dashboard data:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Export compliance report
   * @param {Object} params - Export parameters
   * @param {number} params.days - Number of days to look back
   * @param {string} params.project_id - Optional project ID filter
   * @param {string} params.format - Export format (csv, pdf, xlsx)
   * @returns {Promise<Blob>} Exported report data
   */
  async exportReport(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/dashboard/export`, {
        params,
        responseType: 'blob',
      });
      return response.data;
    } catch (error) {
      console.error('Failed to export compliance report:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get compliance trends data
   * @param {Object} params - Query parameters
   * @returns {Promise<Object>} Trends data
   */
  async getTrendsData(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/trends`, { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch compliance trends:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get task compliance details
   * @param {number} taskId - Task ID
   * @returns {Promise<Object>} Task compliance details
   */
  async getTaskComplianceDetails(taskId) {
    try {
      const response = await axios.get(`${this.baseURL}/tasks/${taskId}`);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch task compliance details:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get user compliance details
   * @param {number} userId - User ID
   * @returns {Promise<Object>} User compliance details
   */
  async getUserComplianceDetails(userId) {
    try {
      const response = await axios.get(`${this.baseURL}/users/${userId}`);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch user compliance details:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get project compliance details
   * @param {number} projectId - Project ID
   * @returns {Promise<Object>} Project compliance details
   */
  async getProjectComplianceDetails(projectId) {
    try {
      const response = await axios.get(`${this.baseURL}/projects/${projectId}`);
      return response.data;
    } catch (error) {
      console.error('Failed to fetch project compliance details:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get security events
   * @param {Object} params - Query parameters
   * @returns {Promise<Array>} Security events
   */
  async getSecurityEvents(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/security-events`, { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch security events:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Resolve security event
   * @param {number} eventId - Event ID
   * @param {Object} data - Resolution data
   * @returns {Promise<Object>} Resolution result
   */
  async resolveSecurityEvent(eventId, data = {}) {
    try {
      const response = await axios.post(`${this.baseURL}/security-events/${eventId}/resolve`, data);
      return response.data;
    } catch (error) {
      console.error('Failed to resolve security event:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get compliance metrics
   * @param {Object} params - Query parameters
   * @returns {Promise<Object>} Compliance metrics
   */
  async getComplianceMetrics(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/metrics`, { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch compliance metrics:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Get compliance alerts
   * @param {Object} params - Query parameters
   * @returns {Promise<Array>} Compliance alerts
   */
  async getComplianceAlerts(params = {}) {
    try {
      const response = await axios.get(`${this.baseURL}/alerts`, { params });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch compliance alerts:', error);
      throw this.handleError(error);
    }
  }

  /**
   * Download compliance report file
   * @param {Blob} blob - Report blob data
   * @param {string} filename - Filename for download
   */
  // eslint-disable-next-line class-methods-use-this
  downloadReport(blob, filename) {
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  }

  /**
   * Format date range for API calls
   * @param {string} range - Date range (last_week, last_month, etc.)
   * @returns {Object} Formatted date range
   */
  // eslint-disable-next-line class-methods-use-this
  formatDateRange(range) {
    const now = new Date();
    const ranges = {
      last_week: 7,
      last_month: 30,
      last_quarter: 90,
      last_year: 365,
    };

    const days = ranges[range] || 30;
    const startDate = new Date(now.getTime() - (days * 24 * 60 * 60 * 1000));

    return {
      start_date: startDate.toISOString().split('T')[0],
      end_date: now.toISOString().split('T')[0],
      days,
    };
  }

  /**
   * Process compliance data for charts
   * @param {Array} data - Raw compliance data
   * @returns {Array} Processed chart data
   */
  processChartData(data) {
    if (!data || !Array.isArray(data)) {
      return [];
    }

    return data.map((item) => ({
      date: new Date(item.date),
      value: item.value || 0,
      count: item.count || 0,
      label: this.formatChartLabel(item.date),
    }));
  }

  /**
   * Format chart label
   * @param {string} dateString - Date string
   * @returns {string} Formatted label
   */
  // eslint-disable-next-line class-methods-use-this
  formatChartLabel(dateString) {
    const date = new Date(dateString);
    const options = {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    };
    return date.toLocaleDateString('en-US', options);
  }

  /**
   * Calculate compliance score
   * @param {Object} metrics - Compliance metrics
   * @returns {number} Compliance score (0-100)
   */
  // eslint-disable-next-line class-methods-use-this
  calculateComplianceScore(metrics) {
    if (!metrics) return 0;

    const weights = {
      success_rate: 0.4,
      security_score: 0.3,
      activity_score: 0.2,
      team_score: 0.1,
    };

    let score = 0;
    Object.keys(weights).forEach((key) => {
      if (metrics[key] !== undefined) {
        score += (metrics[key] / 100) * weights[key];
      }
    });

    return Math.round(score * 100);
  }

  /**
   * Get compliance status color
   * @param {number} score - Compliance score
   * @returns {string} Color class
   */
  // eslint-disable-next-line class-methods-use-this
  getComplianceStatusColor(score) {
    if (score >= 90) return 'success';
    if (score >= 70) return 'warning';
    return 'error';
  }

  /**
   * Get compliance status text
   * @param {number} score - Compliance score
   * @returns {string} Status text
   */
  // eslint-disable-next-line class-methods-use-this
  getComplianceStatusText(score) {
    if (score >= 90) return 'Excellent';
    if (score >= 80) return 'Good';
    if (score >= 70) return 'Fair';
    if (score >= 60) return 'Poor';
    return 'Critical';
  }

  /**
   * Handle API errors
   * @param {Error} error - API error
   * @returns {Error} Processed error
   */
  // eslint-disable-next-line class-methods-use-this
  handleError(error) {
    if (error.response) {
      // Server responded with error status
      const message = error.response.data?.message || error.response.data?.error || 'Server error';
      return new Error(`${error.response.status}: ${message}`);
    }
    if (error.request) {
      // Request was made but no response received
      return new Error('Network error: Unable to connect to server');
    }
    // Something else happened
    return new Error(error.message || 'Unknown error occurred');
  }
}

// Create and export singleton instance
const complianceService = new ComplianceService();
export default complianceService;
