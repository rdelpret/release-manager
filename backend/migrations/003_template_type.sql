-- Add template_type to campaigns (single, soundcloud_flip, lp_ep)
ALTER TABLE campaigns ADD COLUMN template_type TEXT NOT NULL DEFAULT 'single';
