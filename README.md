# Colablists

4. Create a house and add users to it
    - Users share all lists in house.
5. Allow copying list in text format
6. Allow to import a list from text format
1. Add form validation
7. Move create a new list component to a modal, shown when user selects add list, +, or whatevs
2. Add checks to list items
3. Improve list detail edition 
    - User select with search and better UI
    - Allow to select a house on list [ house stuff ]

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

