CREATE TABLE IF NOT EXISTS advertisements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source VARCHAR(50) NOT NULL, -- internal, sponsor, admob
    placement VARCHAR(50) NOT NULL, -- carousel, card_deck, popup_modal, interstitial
    image_url TEXT,
    link TEXT,
    sponsor VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ads_active_order ON advertisements(is_active, sort_order);
CREATE INDEX idx_ads_placement_active ON advertisements(placement, is_active);
