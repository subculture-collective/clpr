#!/bin/sh
# Clipper discovery scraper — runs inside the backend container
# Scrapes popular Twitch broadcasters for clips → discovery_clips table
# Falls back to a seed list when clip_submissions is empty (fresh DB)

BROADCASTERS="xQc,pokimane,caseoh_,Agent00,BotezLive,Maximilian_DOOD,ohnePixel,ExtraEmily,Zoil,Elajjaz,itmeJP,APPLESHAMPOO,Bonnie,Pamaj,ashswag,shindigwow"

# Check if clip_submissions has any data
HAS_SUBMISSIONS=$(docker exec pg17-clustr psql -U clpr -d clpr_db -tAc "SELECT EXISTS(SELECT 1 FROM clip_submissions LIMIT 1);")

if [ "$HAS_SUBMISSIONS" = "t" ]; then
    # Let the scraper discover broadcasters from submissions
    docker exec clpr-backend /root/scraper -min-views 50 -max-age-days 7
else
    # Use seed list for fresh DB
    docker exec clpr-backend /root/scraper -min-views 50 -max-age-days 7 -broadcasters "$BROADCASTERS"
fi
