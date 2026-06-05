# 📼 RUKHSAT — College Farewell Scrapbook & Memory Hub

Welcome to **RUKHSAT**, a highly aesthetic, premium, and interactive college farewell scrapbook website. Designed with a cozy retro mood, this site features linen textures, polaroid card structures, handwritten styles, an interactive cassette tape player, and a fully searchable Google Drive video hub, perfect for alumni to walk down memory lane.

---

## 🚀 How to Run Locally

Because the site is built on premium vanilla technologies (HTML5, Custom CSS variables, and native ES6 JavaScript modules), **there are zero installation scripts, node libraries, or compilation builds needed.**

1. Simply open the folder on your system.
2. Double-click the **`index.html`** file.
3. The website will load instantly in your favorite browser!

---

## 🎬 How to Embed Your Own Google Drive Videos

The site supports seamless video streaming from Google Drive files via theater modals. To upload your own class memories:

### Step 1: Upload and Share the Video
1. Upload your `.mp4` video files to your **Google Drive**.
2. Right-click the video file, select **Share** ➡️ **Share**.
3. Under *General Access*, change the permission settings from **"Restricted"** to **"Anyone with the link can view"** (CRITICAL: If restricted, alumni won't be able to stream the clip!).
4. Click **Copy link** and then **Done**.

### Step 2: Extract the File ID
Look at the copied share link. It will look like this:
```
https://drive.google.com/file/d/1-JptJk9U9hZ916ZixN2t_yLzD0Dkvt6s/view?usp=sharing
```
Your unique **Google Drive File ID** is the long string of alphanumeric characters between `/d/` and `/view`.
In the example above, the ID is: `1-JptJk9U9hZ916ZixN2t_yLzD0Dkvt6s`.

### Step 3: Insert into `videos.js`
Open the **`videos.js`** file in any standard text editor (VS Code, Notepad, etc.) and update the database structure:

```javascript
const VIDEO_DATABASE = [
  {
    id: "YOUR_EXTRACTED_FILE_ID", // Paste your GDrive ID here
    title: "Chai Runs & Main Gate Walks", // Title of the card
    category: "campus", // Category: campus, classroom, fests, hostel, messages
    description: "reliving the 3 AM tea breaks and evening sessions during exams.",
    duration: "2:45" // Displayed length
  },
  // Add as many videos as you'd like!
];
```

---

## 🎨 Visual Design Guide & Settings

### Custom Colors
You can customize the color palette of the website without changing individual classes! Open **`styles.css`** and edit the CSS variables located inside the `:root` block:

```css
:root {
  --bg-linen: #fdfbf7;           /* Main parchment background */
  --sepia-gold: #c2945d;         /* Accent vintage gold */
  --rose-burgundy: #7c3d49;      /* Warm emotive shade */
  --sage-green-light: #eff2ef;   /* Guestbook cork background */
}
```

### Music Soundtrack
The background cassette tape player streams a soft instrumental track. To replace this with your batch's favorite lo-fi remix or nostalgic theme:
1. Open `index.html`.
2. Look for the `<audio>` tag near the top:
   ```html
   <audio id="bg-audio" loop>
     <source src="https://your-custom-track-url.mp3" type="audio/mpeg">
   </audio>
   ```
3. Swap the `src` link with any stable direct audio link (`.mp3` or `.ogg`).

---

## 🌐 Deploying the Website Online (Free!)

To share this digital scrapbook with all alumni across the world, you can host it for free in under 5 minutes:

### Option A: Vercel / Netlify (Recommended)
1. Go to [Vercel](https://vercel.com) or [Netlify](https://netlify.com) and create a free account.
2. Drag and drop the **RUKHSAT** folder directly into their deployment box.
3. Your website is instantly live with a custom shareable URL (e.g. `rukhsat-classof26.vercel.app`)!

### Option B: GitHub Pages
1. Push this folder to a new repository on your GitHub account.
2. Go to **Settings** ➡️ **Pages**.
3. Under *Build and deployment*, set Source to **Deploy from a branch** and select `main` (or `master`) ➡️ `/root`.
4. Click Save, and GitHub will host your site for free!

---

*Made with ❤️ for the graduating Batch of 2026. Keep these memories alive forever.*
