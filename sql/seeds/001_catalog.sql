BEGIN;

INSERT INTO categories (name, slug)
VALUES
    ('AI', 'ai'),
    ('Technology', 'technology'),
    ('Sports', 'sports'),
    ('Entertainment', 'entertainment'),
    ('Business', 'business'),
    ('Science', 'science'),
    ('Gaming', 'gaming')
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name;

INSERT INTO feeds (id, name, url, submitted_by, last_fetched_at)
VALUES
    (gen_random_uuid(), 'OpenAI News', 'https://openai.com/news/rss.xml', NULL, NULL),
    (gen_random_uuid(), 'Google AI Blog', 'https://blog.google/technology/ai/rss/', NULL, NULL),
    (gen_random_uuid(), 'Hugging Face Blog', 'https://huggingface.co/blog/feed.xml', NULL, NULL),
    (gen_random_uuid(), 'MIT Technology Review AI', 'https://www.technologyreview.com/topic/artificial-intelligence/feed/', NULL, NULL),
    (gen_random_uuid(), 'TechCrunch AI', 'https://techcrunch.com/category/artificial-intelligence/feed/', NULL, NULL),
    (gen_random_uuid(), 'Ars Technica', 'https://feeds.arstechnica.com/arstechnica/index', NULL, NULL),
    (gen_random_uuid(), 'The Verge', 'https://www.theverge.com/rss/index.xml', NULL, NULL),
    (gen_random_uuid(), 'WIRED', 'https://www.wired.com/feed/rss', NULL, NULL),
    (gen_random_uuid(), 'TechCrunch', 'https://techcrunch.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'Hacker News', 'https://news.ycombinator.com/rss', NULL, NULL),
    (gen_random_uuid(), 'ESPN Top Headlines', 'https://www.espn.com/espn/rss/news', NULL, NULL),
    (gen_random_uuid(), 'BBC Sport', 'https://feeds.bbci.co.uk/sport/rss.xml', NULL, NULL),
    (gen_random_uuid(), 'The Guardian Sport', 'https://www.theguardian.com/sport/rss', NULL, NULL),
    (gen_random_uuid(), 'CBS Sports Headlines', 'https://www.cbssports.com/rss/headlines/', NULL, NULL),
    (gen_random_uuid(), 'Sky Sports News', 'https://www.skysports.com/rss/12040', NULL, NULL),
    (gen_random_uuid(), 'Variety', 'https://variety.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'The Hollywood Reporter', 'https://www.hollywoodreporter.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'Deadline', 'https://deadline.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'Rolling Stone', 'https://www.rollingstone.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'Billboard', 'https://www.billboard.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'CNBC Top News', 'https://www.cnbc.com/id/10001147/device/rss/rss.html', NULL, NULL),
    (gen_random_uuid(), 'Fortune', 'https://fortune.com/feed/', NULL, NULL),
    (gen_random_uuid(), 'Entrepreneur', 'https://www.entrepreneur.com/latest.rss', NULL, NULL),
    (gen_random_uuid(), 'Business Insider', 'https://feeds.businessinsider.com/custom/all', NULL, NULL),
    (gen_random_uuid(), 'BBC Business', 'https://feeds.bbci.co.uk/news/business/rss.xml', NULL, NULL),
    (gen_random_uuid(), 'ScienceDaily', 'https://www.sciencedaily.com/rss/all.xml', NULL, NULL),
    (gen_random_uuid(), 'NASA', 'https://www.nasa.gov/feed/', NULL, NULL),
    (gen_random_uuid(), 'MIT School of Science', 'https://news.mit.edu/rss/school/science', NULL, NULL),
    (gen_random_uuid(), 'Quanta Magazine', 'https://api.quantamagazine.org/feed/', NULL, NULL),
    (gen_random_uuid(), 'Phys.org', 'https://phys.org/rss-feed/', NULL, NULL),
    (gen_random_uuid(), 'IGN', 'https://feeds.ign.com/ign/all', NULL, NULL),
    (gen_random_uuid(), 'GameSpot', 'https://www.gamespot.com/feeds/mashup/', NULL, NULL),
    (gen_random_uuid(), 'Polygon', 'https://www.polygon.com/rss/index.xml', NULL, NULL),
    (gen_random_uuid(), 'Eurogamer', 'https://www.eurogamer.net/feed', NULL, NULL),
    (gen_random_uuid(), 'Rock Paper Shotgun', 'https://www.rockpapershotgun.com/feed', NULL, NULL)
ON CONFLICT (url) DO UPDATE
SET
    name = EXCLUDED.name,
    updated_at = NOW();

WITH mappings (feed_url, category_slug) AS (
    VALUES
        ('https://openai.com/news/rss.xml', 'ai'),
        ('https://blog.google/technology/ai/rss/', 'ai'),
        ('https://huggingface.co/blog/feed.xml', 'ai'),
        ('https://www.technologyreview.com/topic/artificial-intelligence/feed/', 'ai'),
        ('https://techcrunch.com/category/artificial-intelligence/feed/', 'ai'),
        ('https://feeds.arstechnica.com/arstechnica/index', 'technology'),
        ('https://www.theverge.com/rss/index.xml', 'technology'),
        ('https://www.wired.com/feed/rss', 'technology'),
        ('https://techcrunch.com/feed/', 'technology'),
        ('https://news.ycombinator.com/rss', 'technology'),
        ('https://www.espn.com/espn/rss/news', 'sports'),
        ('https://feeds.bbci.co.uk/sport/rss.xml', 'sports'),
        ('https://www.theguardian.com/sport/rss', 'sports'),
        ('https://www.cbssports.com/rss/headlines/', 'sports'),
        ('https://www.skysports.com/rss/12040', 'sports'),
        ('https://variety.com/feed/', 'entertainment'),
        ('https://www.hollywoodreporter.com/feed/', 'entertainment'),
        ('https://deadline.com/feed/', 'entertainment'),
        ('https://www.rollingstone.com/feed/', 'entertainment'),
        ('https://www.billboard.com/feed/', 'entertainment'),
        ('https://www.cnbc.com/id/10001147/device/rss/rss.html', 'business'),
        ('https://fortune.com/feed/', 'business'),
        ('https://www.entrepreneur.com/latest.rss', 'business'),
        ('https://feeds.businessinsider.com/custom/all', 'business'),
        ('https://feeds.bbci.co.uk/news/business/rss.xml', 'business'),
        ('https://www.sciencedaily.com/rss/all.xml', 'science'),
        ('https://www.nasa.gov/feed/', 'science'),
        ('https://news.mit.edu/rss/school/science', 'science'),
        ('https://api.quantamagazine.org/feed/', 'science'),
        ('https://phys.org/rss-feed/', 'science'),
        ('https://feeds.ign.com/ign/all', 'gaming'),
        ('https://www.gamespot.com/feeds/mashup/', 'gaming'),
        ('https://www.polygon.com/rss/index.xml', 'gaming'),
        ('https://www.eurogamer.net/feed', 'gaming'),
        ('https://www.rockpapershotgun.com/feed', 'gaming')
)
INSERT INTO feed_categories (feed_id, category_id)
SELECT feeds.id, categories.id
FROM mappings
JOIN feeds ON feeds.url = mappings.feed_url
JOIN categories ON categories.slug = mappings.category_slug
ON CONFLICT (feed_id, category_id) DO NOTHING;

COMMIT;
