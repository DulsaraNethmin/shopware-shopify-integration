import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format } from 'date-fns';

// Combine class names with Tailwind
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

// Format dates consistently
export function formatDate(date: string | Date) {
    const d = typeof date === 'string' ? new Date(date) : date;
    return format(d, 'MMM dd, yyyy HH:mm');
}

// Format error messages from API
export function formatError(error: any): string {
    if (typeof error === 'string') {
        return error;
    }

    if (error?.response?.data?.error) {
        return error.response.data.error;
    }

    if (error?.message) {
        return error.message;
    }

    return 'An unknown error occurred';
}

// Get status badge variant based on status
export function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'info' | 'secondary' {
    switch (status.toLowerCase()) {
        case 'active':
        case 'success':
            return 'success';
        case 'inactive':
            return 'secondary';
        case 'pending':
        case 'in_progress':
            return 'warning';
        case 'failed':
            return 'error';
        default:
            return 'info';
    }
}

// Truncate text with ellipsis
export function truncate(text: string, length: number) {
    if (text.length <= length) {
        return text;
    }
    return text.substring(0, length) + '...';
}

// Format JSON for display
export function formatJSON(json: string) {
    try {
        return JSON.stringify(JSON.parse(json), null, 2);
    } catch (e) {
        return json;
    }
}

// Convert camelCase to Title Case
export function titleCase(text: string) {
    // Add space before capital letters and uppercase the first letter
    const result = text.replace(/([A-Z])/g, ' $1');
    return result.charAt(0).toUpperCase() + result.slice(1);
}

export function previewTransformation(
    sourceValue: any,
    transformType: string,
    transformConfig: string
): { result: any, error?: string } {
    try {
        const config = JSON.parse(transformConfig);

        switch (transformType) {
            case 'none':
                return { result: sourceValue };

            case 'format':
                // Date formatting preview
                if (config.sourceFormat && config.destFormat) {
                    try {
                        const date = new Date(sourceValue);
                        return {
                            result: format(date, config.destFormat)
                        };
                    } catch (e: unknown) {
                        const errorMessage = e instanceof Error ? e.message : 'Unknown error';
                        return {
                            result: sourceValue,
                            error: 'Invalid date format'
                        };
                    }
                }
                return {
                    result: sourceValue,
                    error: 'Invalid format configuration'
                };

            case 'convert':
                // Type conversion preview
                if (config.type === 'string') {
                    return { result: String(sourceValue) };
                } else if (config.type === 'int' || config.type === 'integer') {
                    const value = parseInt(sourceValue);
                    if (isNaN(value)) {
                        return { result: 0, error: 'Could not convert to integer' };
                    }
                    return { result: value };
                } else if (config.type === 'float' || config.type === 'number') {
                    const value = parseFloat(sourceValue);
                    if (isNaN(value)) {
                        return { result: 0, error: 'Could not convert to float' };
                    }
                    return { result: value };
                } else if (config.type === 'boolean') {
                    if (typeof sourceValue === 'string') {
                        return { result: sourceValue.toLowerCase() === 'true' };
                    }
                    return { result: Boolean(sourceValue) };
                }
                return {
                    result: sourceValue,
                    error: 'Unknown conversion type'
                };

            case 'map':
                // Value mapping preview
                if (config[sourceValue] !== undefined) {
                    return { result: config[sourceValue] };
                } else if (config._default !== undefined) {
                    return { result: config._default };
                }
                return {
                    result: sourceValue,
                    error: 'No mapping found for this value'
                };

            case 'template':
                // Template-based transformation
                if (config.template) {
                    let result = config.template;
                    // Handle simple value substitution
                    result = result.replace(/\{\{value\}\}/g, String(sourceValue));

                    // Handle other variables if present
                    if (config.variables) {
                        for (const [key, value] of Object.entries(config.variables)) {
                            result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), String(value));
                        }
                    }

                    return { result };
                }
                return {
                    result: sourceValue,
                    error: 'Missing template in configuration'
                };

            case 'graphql_id':
                // GraphQL Global ID transformation
                if (config.resource_type && config.direction) {
                    if (config.direction === 'to_global') {
                        return {
                            result: `gid://shopify/${config.resource_type}/${sourceValue}`
                        };
                    } else if (config.direction === 'from_global') {
                        // Extract ID from global ID
                        const parts = String(sourceValue).split('/');
                        if (parts.length >= 4) {
                            return { result: parts[parts.length - 1] };
                        }
                        return {
                            result: sourceValue,
                            error: 'Invalid global ID format'
                        };
                    }
                }
                return {
                    result: sourceValue,
                    error: 'Invalid GraphQL ID configuration'
                };

            case 'array_map':
                // Array mapping preview (simplified)
                if (!Array.isArray(sourceValue)) {
                    return {
                        result: [],
                        error: 'Source value is not an array'
                    };
                }

                try {
                    const mapped = sourceValue.map(item => {
                        // Extract source field value if specified
                        let sourceFieldValue = item;
                        if (config.source_path) {
                            const pathParts = config.source_path.split('.');
                            let current = item;
                            for (const part of pathParts) {
                                if (current === undefined || current === null) break;
                                current = current[part];
                            }
                            sourceFieldValue = current;
                        }

                        // Map the value if mapping is provided
                        if (config.mapping && config.mapping[sourceFieldValue]) {
                            sourceFieldValue = config.mapping[sourceFieldValue];
                        }

                        // Create destination object with the proper structure
                        if (config.dest_path) {
                            const result = {};
                            let current: Record<string, any> = result;
                            const pathParts = config.dest_path.split('.');
                            for (let i = 0; i < pathParts.length; i++) {
                                const part = pathParts[i];
                                if (i === pathParts.length - 1) {
                                    current[part] = sourceFieldValue;
                                } else {
                                    // Make sure the property exists as an object before trying to access it
                                    if (!current[part] || typeof current[part] !== 'object') {
                                        current[part] = {};
                                    }
                                    current = current[part] as Record<string, any>;
                                }
                            }
                            return result;
                        }

                        return sourceFieldValue;
                    });

                    return { result: mapped };
                } catch (e: unknown) {
                    const errorMessage = e instanceof Error ? e.message : 'Unknown error';
                    return {
                        result: [],
                        error: 'Error mapping array: ' + errorMessage
                    };
                }

            case 'json_path':
                // JSON path extraction preview
                if (!config.path) {
                    return {
                        result: sourceValue,
                        error: 'Missing path in configuration'
                    };
                }

                try {
                    const pathParts = config.path.split('.');
                    let result = sourceValue;

                    for (const part of pathParts) {
                        // Handle array index notation (e.g., items[0])
                        const arrayMatch = part.match(/^([^\[]+)\[(\d+)\]$/);
                        if (arrayMatch) {
                            const [_, key, indexStr] = arrayMatch;
                            const index = parseInt(indexStr);
                            result = result[key][index];
                        } else {
                            result = result[part];
                        }

                        if (result === undefined) {
                            return {
                                result: null,
                                error: `Path segment '${part}' not found`
                            };
                        }
                    }

                    return { result };
                } catch (e: unknown) {
                    const errorMessage = e instanceof Error ? e.message : 'Unknown error';
                    return {
                        result: null,
                        error: 'Error extracting JSON path: ' + errorMessage
                    };
                }

            case 'media_map':
                // Media mapping preview
                if (!Array.isArray(sourceValue)) {
                    return {
                        result: [],
                        error: 'Source value is not an array'
                    };
                }

                try {
                    const baseUrl = config.base_url || '';
                    const mapped = sourceValue.map((media, index) => {
                        return {
                            altText: media.alt || '',
                            mediaContentType: 'IMAGE',
                            originalSource: baseUrl + (media.url || ''),
                            position: index + 1
                        };
                    });

                    return { result: mapped };
                } catch (e: unknown) {
                    const errorMessage = e instanceof Error ? e.message : 'Unknown error';
                    return {
                        result: [],
                        error: 'Error mapping media: ' + errorMessage
                    };
                }

            case 'metafield':
                // Metafield creation preview
                if (!config.namespace || !config.key) {
                    return {
                        result: null,
                        error: 'Missing namespace or key in configuration'
                    };
                }

                return {
                    result: {
                        namespace: config.namespace,
                        key: config.key,
                        value: String(sourceValue),
                        type: config.type || 'string'
                    }
                };

            case 'entity_lookup':
                // Entity lookup preview (simplified, would need backend in real implementation)
                return {
                    result: `[${config.entity_type}:${sourceValue}]`,
                    error: 'Entity lookup requires backend support'
                };

            case 'conditional':
                // Conditional transformation preview
                if (!config.conditions || !Array.isArray(config.conditions)) {
                    return {
                        result: sourceValue,
                        error: 'Invalid conditions in configuration'
                    };
                }

                for (const condition of config.conditions) {
                    let matches = false;

                    // Evaluate the condition
                    if (condition.operator === 'equals') {
                        matches = sourceValue === condition.value;
                    } else if (condition.operator === 'contains') {
                        matches = String(sourceValue).includes(condition.value);
                    } else if (condition.operator === 'greater_than') {
                        matches = sourceValue > condition.value;
                    } else if (condition.operator === 'less_than') {
                        matches = sourceValue < condition.value;
                    }

                    if (matches && condition.result !== undefined) {
                        return { result: condition.result };
                    }
                }

                // Use default if no conditions match
                if (config.default !== undefined) {
                    return { result: config.default };
                }

                return { result: sourceValue };

            default:
                return {
                    result: sourceValue,
                    error: 'Preview not available for this transformation type'
                };
        }
    } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        return {
            result: sourceValue,
            error: 'Invalid transformation configuration: ' + errorMessage
        };
    }
}