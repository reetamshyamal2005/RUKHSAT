/**
 * RUKHSAT - College Farewell Video Database
 * 
 * To add your own videos:
 * 1. Upload your video to Google Drive.
 * 2. Right-click the video, select "Share" -> "Anyone with the link can view".
 * 3. Copy the link. It looks like: https://drive.google.com/file/d/YOUR_FILE_ID/view?usp=sharing
 * 4. Extract the YOUR_FILE_ID portion (a long string of random alphanumeric characters).
 * 5. Paste the ID below in the `id` field.
 * 6. Set the appropriate category: 'campus', 'classroom', 'fests', 'hostel', or 'messages'.
 */

const VIDEO_DATABASE = [
  {
    id: "1MUlyaT3cGfOjOc2Nz0r6bhsha6KyLtqS", // Replace with your Google Drive File ID
    title: "The Campus Walk & Chai Runs",
    category: "campus",
    description: "Capturing the daily pilgrimage to the canteen, late-night tea stalls, and the scenic pathways that felt like home.",
    duration: "2:45"
  },
  {
    id: "1-2qC_sZ-L2hWb-iR9Bv-o2u7hB9dM2n8", // Replace with your Google Drive File ID
    title: "Classroom Pranks & Backbenchers",
    category: "classroom",
    description: "A tribute to the sleepy afternoon lectures, paper plane dogfights, and proxy attendances that kept us afloat.",
    duration: "4:12"
  },
  {
    id: "1-W4M98QW2798e4fN2yX-8P4p7o2Z-m32", // Replace with your Google Drive File ID
    title: "Cult Fest '25: The Final Dance",
    category: "fests",
    description: "Relive the high-octane energy of our last college festival, the band performances, and the dance battles that shook the auditorium.",
    duration: "6:30"
  },
  {
    id: "1-a982F3q82yX7P4p-o2z9823N_xY_z", // Replace with your Google Drive File ID
    title: "Hostel Chronicles: 3 AM Debates",
    category: "hostel",
    description: "The absolute nonsense debated at ungodly hours, the instant noodle experiments, and birthday bumps in the hostel lobby.",
    duration: "3:55"
  },
  {
    id: "1-Z3y8P4p-xY_zN_a982F3q82yX7P4p-o", // Replace with your Google Drive File ID
    title: "Warm Wishes from the Faculty",
    category: "messages",
    description: "Heartwarming words, advice, and blessings from our beloved professors and department heads for our future journey.",
    duration: "8:15"
  },
  {
    id: "1-p82yX7P4p-o2z9823N_xY_z1-Z3y8P4", // Replace with your Google Drive File ID
    title: "The Farewell Prep & Bloopers",
    category: "classroom",
    description: "Behind the scenes of our farewell video shooting. Stutters, laughter, and heavy emotional moments of editing this site.",
    duration: "5:02"
  }
];
