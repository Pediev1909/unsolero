-- Admit `stack` as an editorial content type.
--
-- A stack is a whole set of tools chosen for one kind of business and one
-- budget: what to run, what was deliberately left out and why, and the total
-- per month. It is the shape the recommendation builder already produces,
-- published as an indexable page under /stacks/{slug} where /build itself is
-- noindex.
--
-- The column check was written inline in 000010, so PostgreSQL named it after
-- the table and column. It is replaced rather than dropped: the content type
-- still decides a row's public route, and a value outside the set the domain
-- knows would publish a row nothing resolves.
ALTER TABLE editorial.entries
    DROP CONSTRAINT entries_content_type_check;

ALTER TABLE editorial.entries
    ADD CONSTRAINT entries_content_type_check CHECK (
        content_type IN ('article', 'guide', 'buying_guide', 'comparison', 'stack')
    );
