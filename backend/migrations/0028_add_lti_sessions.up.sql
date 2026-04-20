-- New migration for LTI Deep Linking sessions
CREATE TABLE IF NOT EXISTS lti_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id UUID NOT NULL REFERENCES lti_platforms(id) ON DELETE CASCADE,
    local_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_type TEXT NOT NULL, -- e.g. 'LtiDeepLinkingRequest'
    return_url TEXT,
    data TEXT, -- Opaque data from LMS
    deployment_id TEXT,
    
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
