-- Point published editorial at the versioned illustration filenames.
--
-- The originals were served with a one-year expiry under a stable filename, so
-- every visitor who loaded a page before that was fixed holds the old artwork
-- for a year and a redraw never reaches them. Changing the header stops it
-- recurring; it does not evict what is already cached. Changing the URL does,
-- for everyone, at once.

UPDATE editorial.entries
   SET hero_image_url = regexp_replace(hero_image_url,
       '^/images/(saas-[a-z-]+)\.svg$', '/images/\1-v2.svg'),
       updated_at = now()
 WHERE hero_image_url ~ '^/images/saas-[a-z-]+\.svg$';
