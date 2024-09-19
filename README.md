# Roadmap

1. ✉️  Events
    - [x] Figure out focus/unfocus events
    - [x] Add group event
    - [x] Delete group event
    - [x] Add Item to group event
    - [x] Remove item from group event
    - [x] Edit item
    - [-] Save list: Ask to save before exit (when list has changed) and through button
2. [x] 📁 SessionManager && SqlSessionManager
3. 💅 Improve the UI
4. 🐳 Create docker container
5. ⚙  Create ci/cd pipeline
6. ☁️  Use AWS CloudFormation to deploy the app
7. 🔎 Improve SEO
8. 🎉 Add project to portfolio 


-- Cleaning up

From (Go's sql transaction docs)[https://go.dev/doc/database/execute-transactions]
- remove Rollback error handler from infra
- Call defer Rollback() on all transactions
- Handle all template execute error

-- Bugs

- Fix bug when deleting multiple items (index get crazy)
- Track session timeouts
- Prompt save before leaving list detail and it is dirty
- Persist user color in DB
- Allow to update list name and colaborators
- Only show lists that a person is a colaborator
- Clear lists from liveEditor map when it's idle

64 + 32 + 16 + 8 + 0 + 0 + 0
1    1     1   1   0   0   0



