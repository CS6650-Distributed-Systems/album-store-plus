-- Drop the reviews table first (depends on albums table)
DROP TABLE IF EXISTS reviews;

-- Drop the albums table next (depends on artists table)
DROP TABLE IF EXISTS albums;

-- Drop the artists table last
DROP TABLE IF EXISTS artists;