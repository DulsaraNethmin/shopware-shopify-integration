import Keycloak from 'keycloak-js';

// Keycloak initialization options
export const keycloakConfig = {
    url: process.env.NEXT_PUBLIC_KEYCLOAK_URL || 'https://your-keycloak-url/auth',
    realm: process.env.NEXT_PUBLIC_KEYCLOAK_REALM || 'your-realm',
    clientId: process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT_ID || 'your-client-id'
};

// Initialize Keycloak instance
let keycloak: Keycloak | null = null;

// This function is used to get the Keycloak instance
// It initializes it if it doesn't exist yet
export const getKeycloakInstance = () => {
    if (typeof window !== 'undefined' && !keycloak) {
        keycloak = new Keycloak(keycloakConfig);
    }
    return keycloak;
};

// Function to init Keycloak
export const initKeycloak = async (): Promise<boolean> => {
    const keycloak = getKeycloakInstance();
    if (!keycloak) return false;

    try {
        const authenticated = await keycloak.init({
            onLoad: 'check-sso',
            silentCheckSsoRedirectUri:
                window.location.origin + '/silent-check-sso.html',
            pkceMethod: 'S256', // Use PKCE for security
        });

        // Setup token refresh
        if (authenticated) {
            setupTokenRefresh(keycloak);
        }

        return authenticated;
    } catch (error) {
        console.error('Failed to initialize Keycloak:', error);
        return false;
    }
};

// Function to refresh the token before it expires
const setupTokenRefresh = (keycloak: Keycloak) => {
    // Refresh token 1 minute before it expires
    setInterval(() => {
        keycloak.updateToken(70)
            .then((refreshed) => {
                if (refreshed) {
                    console.log('Token refreshed');
                }
            })
            .catch(() => {
                console.error('Failed to refresh the token, or the session has expired');
            });
    }, 60000); // Check every minute
};

// Function to get the token
export const getToken = () => {
    const keycloak = getKeycloakInstance();
    return keycloak?.token;
};

// Function to login
export const login = () => {
    const keycloak = getKeycloakInstance();
    keycloak?.login();
};

// Function to logout
export const logout = () => {
    const keycloak = getKeycloakInstance();
    keycloak?.logout();
};

// Function to check if the user is authenticated
export const isAuthenticated = () => {
    const keycloak = getKeycloakInstance();
    return !!keycloak?.authenticated;
};

// Function to get user info
export const getUserInfo = () => {
    const keycloak = getKeycloakInstance();
    if (keycloak?.authenticated) {
        return {
            username: keycloak.tokenParsed?.preferred_username,
            email: keycloak.tokenParsed?.email,
            name: keycloak.tokenParsed?.name,
            roles: keycloak.tokenParsed?.realm_access?.roles || []
        };
    }
    return null;
};