# Frontend Architecture

## Overview

The UniPilot frontend is built with Next.js 14+ using the App Router architecture, React, TypeScript, and Tailwind CSS. The application follows a component-based architecture with a focus on reusable UI components, custom hooks for data management, and optimized loading states.

## Architecture Components

- **Next.js App Router**: File-based routing with server and client components
- **React Components**: Functional components with hooks for state management
- **TypeScript**: Type-safe development with comprehensive type definitions
- **Tailwind CSS**: Utility-first CSS framework with custom design tokens
- **shadcn/ui**: Component library built on Radix UI primitives
- **Custom Hooks**: Data fetching and state management abstractions
- **Wails Integration**: Go backend communication via Wails bindings

## Table of Contents

| Document | Description |
|----------|-------------|
| [Component Library](component-library.md) | Reusable UI components, their props, and usage patterns |
| [Design System](design-system.md) | Design tokens, color schemes, typography, and styling conventions |

## Page Components

### Assignments Module

#### `/app/assignments/page.tsx`
- **Type**: Page Component (Client)
- **Purpose**: Main assignments management page with multiple views (Today, Week, Overdue, Exam, Calendar, List)
- **Location**: `frontend/app/assignments/page.tsx`
- **Features**:
  - Tab-based navigation with URL state management
  - Assignment CRUD operations with optimistic updates
  - Real-time assignment filtering and sorting
  - Calendar and list view modes
  - Assignment details modal integration
  - Deep linking support via URL query parameters
- **State Management**: 
  - React hooks (`useState`, `useEffect`) for local component state
  - Custom hooks for data fetching: `useAssignments`, `useTodayAssignments`, `useWeekAssignments`, `useOverdueAssignments`, `useExamAssignments`
  - Mutation hooks: `useUpdateAssignment`, `useDeleteAssignment`, `useCreateAssignment`
  - URL state via `useSearchParams` and `useRouter` from Next.js
- **Component State**:
  - `selectedAssignmentID`: Currently selected assignment for details modal
  - `selectedAssignmentEdit`: Assignment being edited in edit dialog
  - `selectedDate`: Selected date for day assignments modal
- **URL Query Parameters**:
  - `view`: Active tab view ("today" | "week" | "overdue" | "exam" | "calendar" | "list")
  - `course`: Course filter value
  - `status`: Status filter value
  - `priority`: Priority filter value
  - `assignment`: Assignment ID for deep linking to details modal
- **Event Handlers**:
  - `handleAssignmentClick`: Opens assignment details modal
  - `handleEditAssignment`: Updates assignment fields with optimistic updates
  - `handleToggleComplete`: Toggles assignment completion status
  - `handleDeleteAssignment`: Deletes assignment with optimistic UI updates
  - `handleAddAssignment`: Creates new assignment
  - `handleMoveAssignment`: Updates assignment deadline (calendar drag-and-drop)
  - `handleTabChange`: Synchronizes tab selection with URL
- **Interactions**: 
  - Assignment creation, editing, deletion
  - Status toggling (complete/incomplete)
  - Date-based filtering and calendar navigation
  - Assignment detail view on click
  - Calendar drag-and-drop for deadline changes
- **Loading States**: 
  - Shows loading spinner while fetching assignments
  - Displays error message if data fetch fails
  - Uses `loading.tsx` for route transition loading states
- **Toast Notifications**: Success/error feedback for all CRUD operations

#### `/app/assignments/loading.tsx`
- **Type**: Loading Skeleton Component
- **Purpose**: Provides a skeleton loading UI during route transitions for the assignments page
- **Location**: `frontend/app/assignments/loading.tsx`
- **Features**:
  - Automatically displayed by Next.js during route transitions
  - Matches the layout structure of the actual assignments page
  - Includes skeleton placeholders for:
    - Page header (title and description)
    - Filter/search bar card with control skeletons
    - List of 5 assignment card skeletons (checkbox, title, metadata, dates)
- **Styling**: Uses shadcn/ui `Skeleton` component with Tailwind CSS classes
- **Design Patterns**: Glass morphism effects (`glass`, `glass-dark` classes) for visual consistency
- **Props**: None (default export component)
- **Returns**: JSX.Element - A skeleton loading UI matching the assignments page layout

### AI Chat Module

#### `/app/chat/page.tsx`
- **Type**: Page Component (Client)
- **Purpose**: AI-powered chat interface for assignment-specific assistance
- **Location**: `frontend/app/chat/page.tsx`
- **Features**:
  - Assignment-specific AI chat interface
  - Sidebar displaying assignment documents
  - Conversation history persistence
  - Integration with AI SDK for chat functionality
  - Document context for RAG (Retrieval-Augmented Generation)
- **State Management**:
  - `useAssignments` hook to fetch all assignments
  - URL query parameter `assignment` to identify the target assignment
  - Assignment lookup from global assignments state
- **URL Query Parameters**:
  - `assignment`: Required assignment ID (number) to load the chat context
- **Component Structure**:
  - `SidebarProvider`: Wraps the page to provide sidebar context
  - `AiChatSidebar`: Displays assignment title and associated documents
  - `AIChatInterface`: Main chat interface component with message history
- **Loading States**:
  - Shows loading spinner while fetching assignments
  - Displays error message if assignment is not found
- **Error Handling**:
  - Validates assignment ID from URL
  - Checks if assignment exists in fetched data
  - Shows "Assignment not found" error if assignment doesn't exist
- **Styling**: Uses `page-chat` class with animated background elements

### Community Module

#### `/app/community/page.tsx`
- **Type**: Page Component (Client)
- **Purpose**: User discovery and social connections page with tabbed interface
- **Location**: `frontend/app/community/page.tsx`
- **Features**:
  - Tab-based navigation (Explore, Followers, Following)
  - User discovery and search functionality
  - Follow/unfollow user management
  - Responsive grid layout for user cards
- **State Management**:
  - `useAuthContext` hook to access user relationship data
  - Data fetched at provider level for global state access
  - No local state management (stateless component)
- **Component Structure**:
  - `Tabs`: shadcn/ui tab component for navigation
  - `ExploreView`: Displays all users for discovery
  - `FollowersView`: Shows users who follow the current user
  - `FollowingView`: Displays users the current user is following
- **Data Flow**:
  - `followers`: Array of users following the current user
  - `following`: Array of users the current user is following
  - `users`: Array of all users for exploration
- **Tab Views**:
  - `explore`: Default view for discovering new users
  - `followers`: View users who follow you
  - `following`: View users you are following
- **Styling**: Uses glass morphism effects and animated background elements
- **Loading States**: Uses `loading.tsx` for route transition loading states

#### `/app/community/loading.tsx`
- **Type**: Loading Skeleton Component
- **Purpose**: Provides a skeleton loading UI during route transitions for the community page
- **Location**: `frontend/app/community/loading.tsx`
- **Features**:
  - Automatically displayed by Next.js during route transitions
  - Matches the layout structure of the actual community page
  - Includes skeleton placeholders for:
    - Page header (title and description)
    - Tab navigation bar (Explore, Followers, Following tabs)
    - Grid of 6 user card skeletons (avatar, name, bio, action buttons)
- **Styling**: Uses shadcn/ui `Skeleton` component with Tailwind CSS classes
- **Design Patterns**: Glass morphism effects (`glass`, `glass-dark` classes) for visual consistency
- **Layout**: Responsive grid (1 column mobile, 2 columns tablet, 3 columns desktop)
- **Props**: None (default export component)
- **Returns**: JSX.Element - A skeleton loading UI matching the community page layout

### Courses Module

#### `/app/courses/page.tsx`
- **Type**: Page Component (Client)
- **Purpose**: Course management page with schedule and list views
- **Location**: `frontend/app/courses/page.tsx`
- **Features**:
  - Tab-based navigation with URL state management
  - Course CRUD operations with optimistic updates
  - Schedule view (calendar-based) and list view (table-based)
  - Course details modal integration
  - Course sharing via link requests
  - Deep linking support via URL query parameters
- **State Management**:
  - React hooks (`useState`, `useEffect`) for local component state
  - Custom hooks for data fetching: `useCourses`
  - Mutation hooks: `useUpdateCourse`, `useDeleteCourse`, `useCreateCourse`
  - URL state via `useSearchParams` and `useRouter` from Next.js
- **Component State**:
  - `selectedCourseId`: Currently selected course for details modal
  - `selectedDeleteCourseId`: Course ID for delete confirmation dialog
  - `isLinkRequestModalOpen`: Controls link request modal visibility
- **URL Query Parameters**:
  - `view`: Active tab view ("schedule" | "list")
  - `course`: Course code for deep linking to course details
  - `semester`: Semester filter value
  - `instructor`: Instructor filter value
- **Event Handlers**:
  - `handleCourseClick`: Opens course details modal
  - `handleEditCourse`: Updates course fields with optimistic updates
  - `handleDeleteCourse`: Deletes course (triggers confirmation dialog)
  - `handleDeleteCourseClick`: Opens delete confirmation dialog
  - `handleAddCourse`: Creates new course
  - `handleTabChange`: Synchronizes tab selection with URL
- **Interactions**:
  - Course creation, editing, deletion
  - Course detail view on click
  - Course sharing via link requests
  - Schedule and list view switching
- **Loading States**:
  - Shows loading spinner while fetching courses
  - Displays error message if data fetch fails
  - Uses internal loading state (not loading.tsx)
- **Toast Notifications**: Success feedback for course deletion

#### `/app/courses/course-delete-dialog.tsx`
- **Type**: Dialog Component
- **Purpose**: Confirmation dialog for course deletion with impact warning
- **Location**: `frontend/app/courses/course-delete-dialog.tsx`
- **Features**:
  - Displays assignment count that will be deleted
  - Warns about permanent data loss
  - Fetches course assignments to show accurate impact
  - Toast notification on successful deletion
- **Props**:
  - `isOpen`: Controls dialog visibility
  - `onClose`: Callback to close the dialog
  - `courseId`: ID of the course to delete
  - `courses`: Array of all courses to find the target course
  - `onDelete`: Callback to execute deletion
- **Data Fetching**: Uses `useCourseAssignments` hook to fetch assignment count
- **Styling**: Uses glass morphism effect with destructive button variant

#### `/app/courses/loading.tsx`
- **Type**: Loading Component (No-op)
- **Purpose**: Placeholder loading component that returns null
- **Location**: `frontend/app/courses/loading.tsx`
- **Features**:
  - Returns null (no skeleton UI)
  - Satisfies Next.js App Router requirement for loading.tsx
  - Page handles its own loading state internally
- **Returns**: null

### Authentication Module

#### `/app/login/page.tsx`
- **Type**: Page Component (Client)
- **Purpose**: Dedicated login route for user authentication
- **Location**: `frontend/app/login/page.tsx`
- **Features**:
  - Renders LoginForm component with success callback
  - Client-side navigation after successful authentication
  - Redirects to dashboard on login success
- **State Management**:
  - Uses Next.js `useRouter` for client-side navigation
  - No local state management (stateless wrapper component)
- **Component Structure**:
  - `LoginForm`: Reusable login form component from auth components
- **Event Handlers**:
  - `handleLoginSuccess`: Redirects to dashboard after successful login
- **Navigation**:
  - Redirects to `/` (dashboard/home) after successful authentication
  - Uses client-side navigation for smooth transition
- **Styling**: Inherits styling from LoginForm component

## Loading States & UI Skeletons

### Next.js Loading Patterns

The application uses Next.js App Router's built-in loading state management:

- **`loading.tsx` files**: Automatically displayed during route transitions
- **Skeleton Components**: Provide visual placeholders that match actual content layout
- **Perceived Performance**: Improves user experience by showing structure immediately

### Skeleton Component Structure

Loading skeletons follow a consistent pattern:
1. **Header Section**: Title and description placeholders
2. **Filter/Controls Section**: Interactive element skeletons (tabs, filters, search bars)
3. **Content List**: Multiple item skeletons matching the actual list structure
4. **Grid Layouts**: Responsive grid skeletons for card-based layouts (community, courses)

## Component Organization

```
frontend/
├── app/                    # Next.js App Router pages
│   ├── assignments/        # Assignments module
│   │   ├── page.tsx        # Main assignments page
│   │   └── loading.tsx     # Loading skeleton
│   ├── chat/               # AI Chat module
│   │   └── page.tsx        # AI chat page
│   ├── community/          # Community module
│   │   ├── page.tsx        # Main community page
│   │   └── loading.tsx      # Loading skeleton
│   ├── courses/            # Courses module
│   │   ├── page.tsx        # Main courses page
│   │   ├── course-delete-dialog.tsx  # Delete confirmation dialog
│   │   └── loading.tsx     # Loading component (no-op)
│   ├── login/              # Authentication module
│   │   └── page.tsx        # Login page
│   └── ...
├── components/             # Reusable UI components
│   ├── ui/                 # shadcn/ui primitives
│   ├── auth/               # Authentication components
│   │   ├── login-form.tsx
│   │   └── register-form.tsx
│   ├── assignments/        # Assignment-specific components
│   └── ai-chat/            # AI chat components
│       ├── ai-chat-interface.tsx
│       ├── ai-chat-sidebar.tsx
│       └── ai-chat-documents.tsx
├── hooks/                  # Custom React hooks
├── lib/                    # Utility functions
└── types/                  # TypeScript type definitions
```

## Styling Architecture

### Design System
- **Tailwind CSS**: Utility-first CSS framework
- **Custom Classes**: Glass morphism effects (`glass`, `glass-dark`, `bg-animated`)
- **Component Library**: shadcn/ui components with Radix UI primitives
- **Responsive Design**: Mobile-first approach with breakpoint utilities

### Visual Effects
- **Glass Morphism**: Translucent cards with backdrop blur
- **Animated Backgrounds**: Gradient animations and floating elements
- **Skeleton Loaders**: Smooth loading animations using shadcn/ui Skeleton component

## Data Fetching Patterns

### Custom Hooks
- **`useAssignments`**: Fetches all assignments with caching
- **`useUpdateAssignment`**: Optimistic updates for assignment modifications
- **`useCreateAssignment`**: Assignment creation with cache invalidation
- **`useDeleteAssignment`**: Assignment deletion with optimistic UI updates
- **`useCourses`**: Fetches all courses with caching
- **`useUpdateCourse`**: Optimistic updates for course modifications
- **`useCreateCourse`**: Course creation with cache invalidation
- **`useDeleteCourse`**: Course deletion with optimistic UI updates
- **`useCourseAssignments`**: Fetches assignments for a specific course

### State Management
- React Query (TanStack Query) for server state management
- Optimistic updates for immediate UI feedback
- Automatic cache invalidation on mutations
- Error handling with toast notifications

## Routing & Navigation

### URL State Management
- Query parameters for view state (`?view=today`, `?view=schedule`)
- Filter persistence in URL (`?course=X&status=Y`, `?semester=X&instructor=Y`)
- Deep linking support for assignment details (`?assignment=ID`)
- Deep linking support for course details (`?course=CODE`)
- Assignment context for AI chat (`/chat?assignment=ID`)

### Navigation Patterns
- Client-side navigation with Next.js router
- Tab-based navigation with URL synchronization
- Modal overlays for detail views without route changes

