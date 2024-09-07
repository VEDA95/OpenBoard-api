<p align="center">
    <img alt="Open Board Logo" height="128" src="https://raw.githubusercontent.com/VEDA95/OpenBoard-API/main/docs/openboard_logo.png" width="128" />
</p>
<h1 align="center">
    Open Board REST API<br>
    An Open Source REST API Backend<br>
    For Kanban Boards
</h1>

## Overview
This REST API backend for the Open Board project is built using open source tools. The Open Source project serves as an alternative to both
self-hosted, and cloud-hosted solutions for kanban boards.

### TODO LIST FOR INITIAL RELEASE

- Build program for handling DB migrations
- Create Initial DB migrations
- Create authentication system for handling user logins
- Build role/permission system for handling user authorization
- Create websocket server for real-time updates
- Create flows for handling password resets VIA directly by the user, or through the use of forgot password link on the login page
- Add ability to use Identity providers for handling user auth (SSO)
- Implement the ability to group kanban boards by workspace
- Add functionality for defining members and admins for a given kanban board or workspace
- Implement functionality for creating custom fields
- Add functionality for handling file uploads
- Implement the ability to change the positions of cards in a list
- Functionality for commenting and adding attachments to a card
- Build out proper input validation
- Build out API documentation with swagger
- Handle logging activity on cards
- Implement the ability to add checklists to cards
- Implement custom labels for cards
- Implement user sessions