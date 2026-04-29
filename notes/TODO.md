- [ ] make a brew package and investigate other package managers
- [ ] make a demo site
- [x] add screenshots to repo

Bugs:
- [x] duplicate manager modal can bunch up duplicate items and hide entries; add vertical scrolling so all results are visible/reviewable
- [x] initial app load has a long delay before video handling/loading works reliably
  - [x] frontend: separate caching for downloaded videos vs discovered videos
  - [x] backend: add caching layer for video/metadata lookups
  - [x] backend: stream file metadata progressively so UI updates as files are discovered
- [x] adding/updating tags currently triggers a full video refetch; switch to a tag-specific API flow that fetches only videos needing tag updates
- [x] after any select-mode operation completes, automatically disable/exit select mode
- [ ] has trouble with multiple hashes

Tweaks:
- [x] add all files to top of list

Features:
- [x] sort by date uploaded
- [x] new section on the client page: "need to be sorted" or "recently added" 
keeps track of when something was uploaded, if something was uploaded before the last was checked, it will prompt you and give you the option to move it to a folder 
- [x] tag videos: tag videos and search by tag.
- [x] watch collections: ability to watch videos next to each other in some type of grid
- [x] favorite videos:
- [x] the ability to delete videos: make this a super last resort and with lots of ui to stop the user from doing it
- [x] auto detect duplicates with embeddings?

- [x] after mass sort, deactivate "Need sorting"
- [x] when the "+" button is selected, expand corresponding menu in the side nav
- [x] do not allow the root directory to be deleted
- [x] when hovering, the circle for selecting all should be visible. When selected select mode is active
- [x] add the tags we create next to the visaul "mp4" and "new" tags
- [x] the opition to rename files

- [x] Collection tweaks:
  - [x] master controls on collections: (mass play button, mass volume control)
  - [x] change grid width for collections: (4,3,2,1)
  - [x] make a folder a collection: button when clicked give you the option to make a collection 

- [x] Tag tweaks:
  - [x] for tags: instead of checkboxes, do a mass add/remove and an endpoint to do it in bulk, similar to how we do in the mass sort

- [x] status bar for loading
- [ ] thumbnail caching? cahce directory 
- [x] live preview
- [ ] more info on files that dont match types
- [ ] select button is hidden
- [x] Password able to be set in the ui (pin/password)
- [ ] after mass upload, then re render
- [ ] dont open duplucate manager after refresh
- [ ] when dragging, make little icon
- [ ] lock button is visible even when pin is not visible

Analytics:
- [ ] Most watched videos: videos that were watched the most amount of times, ability to sort by most watched

2.0:
- [ ] auto-tagging follow-up: learn from existing tag assignments and suggest/apply tags automatically for newly added videos