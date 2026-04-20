CREATE TABLE IF NOT EXISTS lti_ags_launches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id UUID NOT NULL REFERENCES lti_platforms(id) ON DELETE CASCADE,
    deployment_id TEXT NOT NULL,
    resource_link_id TEXT NOT NULL,
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    local_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lti_sub TEXT NOT NULL,
    lineitem_url TEXT NOT NULL,
    lineitems_url TEXT NOT NULL DEFAULT '',
    scope_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (platform_id, deployment_id, resource_link_id, local_user_id)
);

CREATE INDEX IF NOT EXISTS idx_lti_ags_launches_exam_id
    ON lti_ags_launches (exam_id, local_user_id);

CREATE INDEX IF NOT EXISTS idx_lti_ags_launches_platform_resource
    ON lti_ags_launches (platform_id, deployment_id, resource_link_id);
