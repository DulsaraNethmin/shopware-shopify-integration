import axios from 'axios';
import { getToken } from './keycloak';

// API client configuration
const apiClient = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Add an interceptor to automatically add the Keycloak token
apiClient.interceptors.request.use(
    (config) => {
        const token = getToken();
        if (token) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Set auth token for API requests (for backward compatibility)
export const setAuthToken = (token: string) => {
    if (token) {
        apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
        delete apiClient.defaults.headers.common['Authorization'];
    }
};

// Type definitions for API responses
export interface Connector {
    id: number;
    name: string;
    type: 'shopware' | 'shopify';
    url: string;
    username?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface FieldMapping {
    id: number;
    dataflow_id: number;
    source_field: string;
    dest_field: string;
    is_required: boolean;
    default_value: string;
    transform_type: 'none' | 'format' | 'convert' | 'map' | 'template';
    transform_config: string;
    created_at: string;
    updated_at: string;
}

export interface Dataflow {
    id: number;
    name: string;
    description: string;
    type: 'product' | 'order';
    status: 'active' | 'inactive';
    source_connector_id: number;
    dest_connector_id: number;
    source_connector: Connector;
    dest_connector: Connector;
    created_at: string;
    updated_at: string;
}

export interface MigrationLog {
    id: number;
    dataflow_id: number;
    status: 'pending' | 'in_progress' | 'success' | 'failed';
    source_identifier: string;
    dest_identifier: string;
    execution_arn: string;
    error_message: string;
    completed_at?: string;
    created_at: string;
    updated_at: string;
}

// API service functions
const api = {
    // Health check for API
    checkHealth: () => apiClient.get('/health'),

    // Connectors
    getConnectors: () => apiClient.get<{ data: Connector[] }>('/connectors'),
    getConnector: (id: number) => apiClient.get<{ data: Connector }>(`/connectors/${id}`),
    createConnector: (connector: Partial<Connector>) => apiClient.post<{ data: Connector; message: string }>('/connectors', connector),
    updateConnector: (id: number, connector: Partial<Connector>) => apiClient.put<{ data: Connector; message: string }>(`/connectors/${id}`, connector),
    deleteConnector: (id: number) => apiClient.delete<{ message: string }>(`/connectors/${id}`),
    testConnection: (id: number) => apiClient.get<{ message: string }>(`/connectors/${id}/test`),

    // Dataflows
    getDataflows: () => apiClient.get<{ data: Dataflow[] }>('/dataflows'),
    getDataflow: (id: number) => apiClient.get<{ data: Dataflow }>(`/dataflows/${id}`),
    createDataflow: (dataflow: Partial<Dataflow>) => apiClient.post<{ data: Dataflow; message: string }>('/dataflows', dataflow),
    updateDataflow: (id: number, dataflow: Partial<Dataflow>) => apiClient.put<{ data: Dataflow; message: string }>(`/dataflows/${id}`, dataflow),
    deleteDataflow: (id: number) => apiClient.delete<{ message: string }>(`/dataflows/${id}`),

    // Field Mappings
    getFieldMappings: (dataflowId: number) => apiClient.get<{ data: FieldMapping[] }>(`/dataflows/${dataflowId}/mappings`),
    createFieldMapping: (dataflowId: number, mapping: Partial<FieldMapping>) =>
        apiClient.post<{ data: FieldMapping; message: string }>(`/dataflows/${dataflowId}/mappings`, mapping),
    updateFieldMapping: (dataflowId: number, mappingId: number, mapping: Partial<FieldMapping>) =>
        apiClient.put<{ data: FieldMapping; message: string }>(`/dataflows/${dataflowId}/mappings/${mappingId}`, mapping),
    deleteFieldMapping: (dataflowId: number, mappingId: number) =>
        apiClient.delete<{ message: string }>(`/dataflows/${dataflowId}/mappings/${mappingId}`),

    // Migration Logs
    getMigrationLogs: (dataflowId: number) => apiClient.get<{ data: MigrationLog[] }>(`/dataflows/${dataflowId}/logs`),
    getMigrationLog: (dataflowId: number, logId: number) => apiClient.get<{ data: MigrationLog }>(`/dataflows/${dataflowId}/logs/${logId}`),

    // Products (for testing)
    getProducts: (connectorId: number) => apiClient.get<{ data: any }>(`/connectors/${connectorId}/products`),
    getProduct: (connectorId: number, productId: string) => apiClient.get<{ data: any }>(`/connectors/${connectorId}/products/${productId}`),

    applyDefaultMappings: (dataflowId: number) => apiClient.post<{ message: string, count: number }>(`/dataflows/${dataflowId}/mappings/defaults`),
};

export default api;