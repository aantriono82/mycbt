CREATE TABLE IF NOT EXISTS lti_platforms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    issuer TEXT NOT NULL UNIQUE,
    client_id TEXT NOT NULL,
    deployment_id TEXT NOT NULL,
    oidc_auth_url TEXT NOT NULL,
    oidc_token_url TEXT NOT NULL,
    jwks_url TEXT NOT NULL,
    
    tool_private_key TEXT NOT NULL,
    tool_public_key TEXT NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lti_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id UUID NOT NULL REFERENCES lti_platforms(id) ON DELETE CASCADE,
    local_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lti_sub TEXT NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE(platform_id, lti_sub)
);

CREATE TABLE IF NOT EXISTS lti_nonces (
    nonce TEXT PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
