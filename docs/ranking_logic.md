# Ranking & Matching System Logic

## 1. Score Calculation (GORM / SQL)

```sql
SELECT 
    u.*,
    (
        1.0 -- base_score
        + (CASE WHEN EXISTS (SELECT 1 FROM user_boosts ub WHERE ub.user_id = u.id AND ub.expired_at > NOW()) THEN (SELECT (value->>0)::FLOAT FROM app_configs WHERE key = 'boost_multiplier') ELSE 0 END) -- boost_score
        + (CASE WHEN EXISTS (SELECT 1 FROM user_subscriptions us JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id WHERE us.user_id = u.id AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = 'priority_likes') THEN (SELECT (spf.value->>0)::FLOAT FROM user_subscriptions us JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id WHERE us.user_id = u.id AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = 'priority_likes' LIMIT 1) ELSE 0 END) -- subscription_score
        + (CASE WHEN u.last_active_at > NOW() - INTERVAL '24 hours' THEN 1.0 ELSE 0.5 END) -- activity_score
        + RANDOM() -- randomness (0 to 1)
    ) as ranking_score
FROM users u
WHERE u.id != ? -- target user
  AND u.status = 'active'
  AND u.id NOT IN (SELECT swiped_id FROM swipes WHERE swiper_id = ?) -- not swiped
  AND u.id NOT IN (SELECT shown_user_id FROM user_impressions WHERE viewer_id = ? AND shown_at > NOW() - INTERVAL '1 hour') -- impression cooldown
ORDER BY ranking_score DESC
LIMIT 20;
```

## 2. Like Queue Processing

When User A likes User B:

```sql
INSERT INTO swipes (swiper_id, swiped_id, direction, priority_score, ranking_score)
VALUES (
    ?, ?, ?, 
    (CASE 
        WHEN ? = 'CRUSH' THEN 100 
        WHEN EXISTS (SELECT 1 FROM user_subscriptions us JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id WHERE us.user_id = ? AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = 'priority_likes') THEN 50
        ELSE 0 
    END),
    ? -- precomputed ranking_score at swipe time
);
```

## 3. Incoming Like Discovery Delay

```sql
SELECT s.*
FROM swipes s
JOIN users u ON s.swiper_id = u.id
WHERE s.swiped_id = ? -- current user (receiver of like)
  AND s.direction IN ('LIKE', 'CRUSH')
  AND s.updated_at < NOW() - (
      SELECT INTERVAL '1 minute' * (value->>0)::INT 
      FROM app_configs 
      WHERE key = (CASE 
          WHEN EXISTS (SELECT 1 FROM user_subscriptions us JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id WHERE us.user_id = ? AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = 'priority_likes') THEN 'incoming_like_delay_premium' 
          ELSE 'incoming_like_delay_free' 
      END)
  )
ORDER BY s.priority_score DESC, s.created_at ASC;
```

## 4. Feature Check Helper (Go)

```go
func (s *SubscriptionUseCase) HasFeature(ctx context.Context, userID uuid.UUID, featureKey string) (bool, interface{}, error) {
    var feature struct {
        Value interface{}
    }
    
    err := s.repo.DB.Table("user_subscriptions us").
        Select("spf.value").
        Joins("JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id").
        Where("us.user_id = ? AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = ?", userID, featureKey).
        Scan(&feature).Error
        
    if err != nil {
        return false, nil, err
    }
    
    if feature.Value == nil {
        return false, nil, nil
    }
    
    return true, feature.Value, nil
}
```

## 5. Anti-Cheat Logic

### Swipe Rate Limit (Redis)
```go
func (s *SwipeUseCase) CheckSpam(userID uuid.UUID) error {
    key := fmt.Sprintf("swipe_limit:%s:%d", userID, time.Now().Hour())
    count, _ := s.redis.Incr(key).Result()
    if count > 100 { // Configurable via DB config
        return errors.New("rate limit exceeded")
    }
    return nil
}
```

### Bot Detection (Heuristic)
- Match-to-Like ratio > 90% in short period.
- Zero replies to matches.
- Abnormal geolocation jumps.
