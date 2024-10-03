# Colablists
This app was built as a learning project. So it has some rough edges.
It's designed to be a collaboratively list builder, where multiple people
can update the list simultenously. 

It's using the following stack:
- Go lang
- HTMX
- Alpine.js
- Sqlite3

## Future roadmap:

- [x] Real-time update of lists.
    - [ ] and marking items as gathered
- [ ] Explore delivery automation
- [ ] Explore diet planning
- [ ] Production ready check-list
    - [ ] Recaptcha in strategic forms
    - [ ] IP request rate limiting
    - [ ] Compressing responses (e.g. html is currently sent with lots of spaces)
