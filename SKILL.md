---
name: farewell-website-design
summary: Create an aesthetic, nostalgic college farewell website with Google Drive video playback and easy navigation.
description: "Use this skill to guide the creation of a college farewell website with a warm nostalgic mood, polished navigation, and Google Drive video embedding."
---

# Farewell Website Design Skill

## What this skill does

This skill helps you build a college farewell website that feels nostalgic, elegant, and easy to use. It walks through the design decisions, layout, content organization, and Google Drive video embedding needed to deliver a polished alumni memory site.

## Workflow

1. Define the experience
   - Choose a nostalgic tone: retro, scrapbook, polaroid, cassette, warm paper, or campus twilight.
   - Identify the main goals: alumni video hub, memory gallery, guestbook, event highlights.
   - Decide on core pages or sections: home, videos, memories, about, contact.

2. Build the structure
   - Create a clear top navigation or sticky menu.
   - Use a hero section with a memorable headline and atmosphere.
   - Arrange content in cards, polaroids, or layered paper panels.
   - Add decorative accents like texture, soft shadows, vintage fonts, and subtle motion.

3. Add the video hub
   - Collect Google Drive share links for alumni videos.
   - Extract Drive file IDs and store them in a structured `videos.js` list.
   - Render video cards with title, category, description, and play button.
   - Open videos in a modal or embedded player for easy viewing.

4. Make it accessible and responsive
   - Keep text contrast high and controls easy to tap.
   - Ensure navigation is visible on desktop and mobile.
   - Use responsive layout patterns for cards, grids, and media.
   - Add keyboard focus states and screen reader friendly labels.

5. Polish and deploy
   - Fine-tune typography, spacing, and color palette.
   - Add background music or ambient sound carefully with play/pause controls.
   - Test playback for each Google Drive video link.
   - Deploy using GitHub Pages, Vercel, Netlify, or simple static hosting.

## Quality checklist

- [ ] Nostalgic visual identity is consistent across pages
- [ ] Navigation is simple and always available
- [ ] Videos load from Google Drive and play in the site experience
- [ ] Mobile layout adapts gracefully to smaller screens
- [ ] Content is readable and controls are accessible
- [ ] Deployment instructions are included for sharing the site

## Example prompts to try

- "Create a nostalgic homepage layout for a college farewell website."
- "Help me embed Google Drive videos into the alumni memory page."
- "Suggest a color palette and typography for a vintage farewell scrapbook site."
- "Write the `videos.js` structure for a Google Drive video gallery."

## Next customizations

- Add a `farewell-video-gallery.prompt.md` for video content creation tasks.
- Create an `applyTo` instruction file for HTML/CSS/JS styling patterns.
- Add a `.instructions.md` file to enforce a nostalgic design checklist across the project.
