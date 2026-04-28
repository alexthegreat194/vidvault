- [ ] make a brew package and investigate other package managers
- [ ] make a demo site
- [x] add screenshots to repo

Bugs:
- [x] duplicate manager modal can bunch up duplicate items and hide entries; add vertical scrolling so all results are visible/reviewable
- [x] initial app load has a long delay before video handling/loading works reliably
  - [x] frontend: separate caching for downloaded videos vs discovered videos
  - [x] backend: add caching layer for video/metadata lookups
  - [x] backend: stream file metadata progressively so UI updates as files are discovered
- [ ] adding/updating tags currently triggers a full video refetch; switch to a tag-specific API flow that fetches only videos needing tag updates
- [ ] after any select-mode operation completes, automatically disable/exit select mode

Tweaks:
- [x] add all files to top of list

Features:
- [x] sort by date uploaded
- [ ] new section on the client page: "need to be sorted" or "recently added" 
keeps track of when something was uploaded, if something was uploaded before the last was checked, it will prompt you and give you the option to move it to a folder 
- [x] tag videos: tag videos and search by tag.
- [ ] auto-tagging follow-up: learn from existing tag assignments and suggest/apply tags automatically for newly added videos
- [ ] watch collections: ability to watch videos next to each other in some type of grid
- [x] favorite videos:
- [x] the ability to delete videos: make this a super last resort and with lots of ui to stop the user from doing it
- [x] auto detect duplicates with embeddings?
- [ ] videos that were watched the most amount of times, ability to sort by most watched