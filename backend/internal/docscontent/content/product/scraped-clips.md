# Imported clips and discovery

Imported and community-submitted clips use the main feed and clip-detail
experience. The old standalone discovery pages have been removed:
`/discover` and `/discover/scraped` redirect to `/`.

The maintained frontend entry points are
[HomePage](../../frontend/src/pages/HomePage.tsx),
[ClipFeed](../../frontend/src/components/clip/ClipFeed.tsx), and
[ClipDetailPage](../../frontend/src/pages/ClipDetailPage.tsx).
Source identity and submission validation are handled by the backend services;
removing the obsolete pages does not remove stored clips or migrations.

See the [clip submission contract](../features/clip-submission-rate-limiting.md)
and [testing guide](../testing/TESTING.md) for validation and coverage.
