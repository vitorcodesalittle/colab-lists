# Colablists

!! TODO:
1. Add form validation
2. Add Favicon
3. Add title

An open-source list manager website to help people doing groceries together.

Features:

- Real-time update of lists.
    - 🍏 Adding items, 
    - 🔴 and marking items as gathered
- Share lists

It works as a standalone web server that persists list content on a simple
sqlite database, handles real-time multi-user list edition in-memory

I started it is a learning project, to get experience with Go and htmx.

But I am slowly improving both the UX edges and scalability issues
as best as I can.

-- Bugs
- Track session timeouts
- Prompt save before leaving list detail and it is dirty
- Persist user color in DB
- Allow to update list name and colaborators
- Only show lists that a person is a colaborator
- Clear lists from liveEditor map when it's idle

