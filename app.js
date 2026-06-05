/**
 * RUKHSAT — College Farewell Core Interactivity
 * Integrates music playback, Google Drive video display,
 * dynamic search/filtering, and persistent scrapbook guestbook board.
 */

document.addEventListener('DOMContentLoaded', () => {


  // ==========================================
  // 2. AUDIO & COMPACT RETRO TAPE CONTROLLER
  // ==========================================
  const audio = document.getElementById('bg-audio');
  const playBtn = document.getElementById('play-btn');
  const muteBtn = document.getElementById('mute-btn');
  const quickMusicBtn = document.getElementById('music-quick-btn');
  const quickMusicIcon = quickMusicBtn.querySelector('.music-icon');
  const quickMusicText = quickMusicBtn.querySelector('.music-status-text');
  
  const playSvg = playBtn.querySelector('.play-svg');
  const pauseSvg = playBtn.querySelector('.pause-svg');
  const muteSvg = muteBtn.querySelector('.mute-svg');
  const unmuteSvg = muteBtn.querySelector('.unmute-svg');
  const tickerTextEl = document.querySelector('.ticker-text');
  
  const reels = document.querySelectorAll('.reel');
  const trackTime = document.querySelector('.track-time');
  const cassetteWidget = document.getElementById('cassette-player');
  const toggleWidgetBtn = document.getElementById('toggle-widget-btn');
  const closePlayerBtn = document.getElementById('close-player-btn');

  let isPlaying = false;
  let isMuted = false;

  // Format audio duration
  function formatTime(seconds) {
    const min = Math.floor(seconds / 60);
    const sec = Math.floor(seconds % 60);
    return `${min}:${sec < 10 ? '0' : ''}${sec}`;
  }

  function showPlayerControls() {
    cassetteWidget.classList.remove('hidden');
    cassetteWidget.classList.remove('collapsed');
  }

  function hidePlayerControls() {
    setMusicState(false);
    cassetteWidget.classList.add('hidden');
  }

  // Update cassette display time
  audio.addEventListener('timeupdate', () => {
    trackTime.textContent = formatTime(audio.currentTime);
  });

  // Display track title from `data-track-title` if present
  try {
    if (audio && audio.dataset && audio.dataset.trackTitle && tickerTextEl) {
      tickerTextEl.textContent = `Play Now: ${audio.dataset.trackTitle}`;
    }
  } catch (e) {
    console.warn('Could not set ticker text', e);
  }

  // Audio load error handling
  if (audio) {
    audio.addEventListener('error', (e) => {
      console.warn('Audio failed to load or play', e);
      if (quickMusicText) quickMusicText.textContent = 'Audio unavailable';
      if (tickerTextEl) tickerTextEl.textContent = 'Audio unavailable';
    });
  }

  // Update soundtrack UI
  function updateMusicUI(playing) {
    isPlaying = playing;

    if (playing) {
      playSvg.classList.add('hidden');
      pauseSvg.classList.remove('hidden');
      reels.forEach(reel => reel.classList.add('reel-playing'));
      quickMusicIcon.classList.add('playing');
      quickMusicText.textContent = 'Pause Soundtrack';
      if (tickerTextEl) {
        tickerTextEl.textContent = `Now Playing: ${audio.dataset.trackTitle || 'Woh Din'}`;
      }
      return;
    }

    playSvg.classList.remove('hidden');
    pauseSvg.classList.add('hidden');
    reels.forEach(reel => reel.classList.remove('reel-playing'));
    quickMusicIcon.classList.remove('playing');
    quickMusicText.textContent = 'Play Soundtrack';
    if (tickerTextEl) {
      tickerTextEl.textContent = `Play Now: ${audio.dataset.trackTitle || 'Woh Din'}`;
    }
  }

  function setMusicState(playing) {
    updateMusicUI(playing);

    if (playing) {
      showPlayerControls();
      audio.play().catch(err => {
        console.warn('Audio autoplay blocked by browser policy. Interaction required.', err);
        updateMusicUI(false);
      });
      return;
    }

    audio.pause();
  }

  // Toggle Playback
  function toggleMusic() {
    setMusicState(!isPlaying);
  }

  audio.addEventListener('play', () => setMusicState(true));
  audio.addEventListener('pause', () => setMusicState(false));

  // Toggle Mute
  function toggleMute() {
    if (isMuted) {
      audio.muted = false;
      isMuted = false;
      unmuteSvg.classList.remove('hidden');
      muteSvg.classList.add('hidden');
    } else {
      audio.muted = true;
      isMuted = true;
      unmuteSvg.classList.add('hidden');
      muteSvg.classList.remove('hidden');
    }
  }

  // Event Listeners
  playBtn.addEventListener('click', toggleMusic);
  quickMusicBtn.addEventListener('click', toggleMusic);
  muteBtn.addEventListener('click', toggleMute);
  closePlayerBtn.addEventListener('click', hidePlayerControls);

  // Toggle Cassette Player Visibility
  toggleWidgetBtn.addEventListener('click', () => {
    cassetteWidget.classList.toggle('collapsed');
  });

  // ==========================================
  // 1. HERO TYPING ANIMATION
  // ==========================================
  const heroTypingText = document.getElementById('hero-typing-text');

  if (heroTypingText) {
    const heroMessages = [
      'A tapestry of late nights, shared laughs, and memories that became family.',
      'Every little moment here is a page from our favorite chapter.',
    ];

    let messageIndex = 0;
    let currentCharIndex = 0;
    let isDeleting = false;

    const typingSpeed = 45;
    const deletingSpeed = 30;
    const pauseBetweenMessages = 1800;

    function animateHeroText() {
      const currentMessage = heroMessages[messageIndex];

      if (!isDeleting) {
        heroTypingText.textContent = currentMessage.slice(0, currentCharIndex + 1);
        currentCharIndex += 1;

        if (currentCharIndex === currentMessage.length) {
          isDeleting = true;
          setTimeout(animateHeroText, pauseBetweenMessages);
          return;
        }
      } else {
        heroTypingText.textContent = currentMessage.slice(0, currentCharIndex - 1);
        currentCharIndex -= 1;

        if (currentCharIndex === 0) {
          isDeleting = false;
          messageIndex = (messageIndex + 1) % heroMessages.length;
        }
      }

      setTimeout(animateHeroText, isDeleting ? deletingSpeed : typingSpeed);
    }

    animateHeroText();
  }

  // Sticky Navbar Shrink Effect
  window.addEventListener('scroll', () => {
    const navbar = document.querySelector('.navbar');
    const navLinks = document.querySelectorAll('.nav-links a');
    const sections = document.querySelectorAll('section');
    
    if (window.scrollY > 50) {
      navbar.classList.add('scrolled');
    } else {
      navbar.classList.remove('scrolled');
    }

    // Dynamic Navigation Highlighter
    let current = '';
    sections.forEach(section => {
      const sectionTop = section.offsetTop - 120;
      if (window.scrollY >= sectionTop) {
        current = section.getAttribute('id');
      }
    });

    navLinks.forEach(link => {
      link.classList.remove('active');
      if (link.getAttribute('href').slice(1) === current) {
        link.classList.add('active');
      }
    });
  });

  // ==========================================
  // 3. VIDEO LOUNGE RENDERING & FILTERS
  // ==========================================
  const gridContainer = document.getElementById('video-grid-container');
  const searchInput = document.getElementById('video-search');
  const filterBtns = document.querySelectorAll('.filter-btn');
  
  // Custom aesthetic SVGs representing categories for visual excellence
  const CATEGORY_ILLUSTRATIONS = {
    campus: `
      <svg viewBox="0 0 100 100" class="retro-doodle">
        <rect x="10" y="30" width="80" height="50" rx="3" fill="#6b7a67" opacity="0.3"/>
        <polygon points="50,10 5,30 95,30" fill="#7c3d49" opacity="0.8"/>
        <line x1="50" y1="30" x2="50" y2="80" stroke="#fcfbf7" stroke-width="2"/>
        <rect x="25" y="45" width="12" height="12" rx="1" fill="#fff" opacity="0.7"/>
        <rect x="63" y="45" width="12" height="12" rx="1" fill="#fff" opacity="0.7"/>
        <path d="M44,80 L44,68 L56,68 L56,80 Z" fill="#3a2518"/>
      </svg>`,
    classroom: `
      <svg viewBox="0 0 100 100" class="retro-doodle">
        <rect x="15" y="20" width="70" height="42" fill="#2d2926" stroke="#c2945d" stroke-width="3"/>
        <line x1="25" y1="28" x2="75" y2="28" stroke="#c2945d" stroke-width="1.5" stroke-dasharray="3"/>
        <text x="32" y="44" fill="#c2945d" font-family="Caveat" font-size="14" font-weight="bold">1+1 = Memories</text>
        <path d="M40,75 L15,92 M60,75 L85,92" stroke="#2d2926" stroke-width="3" stroke-linecap="round"/>
        <rect x="35" y="62" width="30" height="15" fill="#c2945d" rx="2"/>
      </svg>`,
    fests: `
      <svg viewBox="0 0 100 100" class="retro-doodle">
        <circle cx="50" cy="50" r="38" fill="#7c3d49" opacity="0.2"/>
        <path d="M25,65 Q35,30 50,65 T75,65" fill="none" stroke="#7c3d49" stroke-width="4" stroke-linecap="round"/>
        <circle cx="25" cy="65" r="5" fill="#c2945d"/>
        <circle cx="50" cy="65" r="5" fill="#c2945d"/>
        <circle cx="75" cy="65" r="5" fill="#c2945d"/>
        <path d="M20,30 L30,40 L40,30 L50,40 L60,30 L70,40 L80,30" fill="none" stroke="#6b7a67" stroke-width="3"/>
      </svg>`,
    hostel: `
      <svg viewBox="0 0 100 100" class="retro-doodle">
        <rect x="20" y="25" width="60" height="50" rx="4" fill="none" stroke="#2d2926" stroke-width="3"/>
        <line x1="20" y1="50" x2="80" y2="50" stroke="#2d2926" stroke-width="3"/>
        <path d="M30,35 H70 M30,60 H70" stroke="#6b7a67" stroke-width="2" stroke-linecap="round"/>
        <circle cx="35" cy="42" r="3" fill="#7c3d49"/>
        <circle cx="65" cy="67" r="3" fill="#c2945d"/>
      </svg>`,
    messages: `
      <svg viewBox="0 0 100 100" class="retro-doodle">
        <path d="M10,25 L90,25 L50,55 Z" fill="#f4efdf" stroke="#2d2926" stroke-width="3" stroke-linejoin="round"/>
        <rect x="10" y="25" width="80" height="50" fill="none" stroke="#2d2926" stroke-width="3" stroke-linejoin="round"/>
        <path d="M10,75 L42,48 M90,75 L58,48" stroke="#2d2926" stroke-width="3"/>
        <path d="M50,40 Q40,30 50,20 Q60,30 50,40 Z" fill="#7c3d49"/>
      </svg>`
  };

  // Render video list
  function renderVideos(videosArray) {
    gridContainer.innerHTML = '';
    
    if (videosArray.length === 0) {
      gridContainer.innerHTML = `
        <div class="no-results-message">
          <p class="cursive- scribble" style="font-size: 2.2rem; text-align: center; grid-column: 1 / -1; width: 100%; color: var(--rose-burgundy);">
            "No matching reels found..."
          </p>
        </div>`;
      return;
    }

    videosArray.forEach((video, index) => {
      const card = document.createElement('div');
      card.className = 'polaroid-card';
      card.setAttribute('data-id', video.id);
      card.setAttribute('data-title', video.title);
      card.setAttribute('data-desc', video.description);
      card.setAttribute('data-category', video.category);

      // We generate beautiful styled gradient backdrops incorporating vector icons 
      // instead of standard placeholders. Extremely professional and fits the vintage scrapbook aesthetic.
      const categoryVector = CATEGORY_ILLUSTRATIONS[video.category] || CATEGORY_ILLUSTRATIONS.campus;
      
      let gradientStyle = 'linear-gradient(135deg, #efe8d4 0%, #dbcaa0 100%)';
      if (video.category === 'fests') gradientStyle = 'linear-gradient(135deg, #f5eef0 0%, #d5bdc2 100%)';
      if (video.category === 'classroom') gradientStyle = 'linear-gradient(135deg, #f3eed7 0%, #c4b998 100%)';
      if (video.category === 'hostel') gradientStyle = 'linear-gradient(135deg, #fcfbf7 0%, #dbdcd0 100%)';
      if (video.category === 'messages') gradientStyle = 'linear-gradient(135deg, #eaeae0 0%, #bfcab9 100%)';

      card.innerHTML = `
        <div class="tape-sticker" style="top: -14px; left: 50%; transform: translateX(-50%) rotate(${index % 2 === 0 ? '-2' : '2'}deg); width: 85px; height: 26px;"></div>
        <div class="video-thumbnail-container" style="background: ${gradientStyle}">
          ${categoryVector}
          <div class="play-overlay">
            <div class="play-icon-circle">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3"></polygon>
              </svg>
            </div>
          </div>
          <span class="video-length-badge">${video.duration}</span>
        </div>
        <div class="polaroid-body">
          <span class="polaroid-cat-tag">${video.category}</span>
          <h4 class="polaroid-card-title">${video.title}</h4>
          <p class="polaroid-card-desc">${video.description}</p>
        </div>
      `;

      // Click event for modal trigger
      card.addEventListener('click', () => {
        openVideoModal(video);
      });

      gridContainer.appendChild(card);
    });
  }

  // Initial Video Load
  renderVideos(VIDEO_DATABASE);

  // Search filter trigger
  function handleFilterChange() {
    const searchQuery = searchInput.value.toLowerCase();
    const activeCategoryBtn = document.querySelector('.filter-btn.active');
    const selectedCategory = activeCategoryBtn.getAttribute('data-category');

    const filteredList = VIDEO_DATABASE.filter(video => {
      const matchesSearch = video.title.toLowerCase().includes(searchQuery) || 
                            video.description.toLowerCase().includes(searchQuery);
      const matchesCategory = selectedCategory === 'all' || video.category === selectedCategory;
      
      return matchesSearch && matchesCategory;
    });

    renderVideos(filteredList);
  }

  searchInput.addEventListener('input', handleFilterChange);

  filterBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      filterBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      handleFilterChange();
    });
  });

  // Toggle Video Settings Instructions panel
  const instructionsBtn = document.getElementById('instructions-btn');
  const instructionsContent = document.getElementById('instructions-content');

  if (instructionsBtn && instructionsContent) {
    const instructionsArrow = instructionsBtn.querySelector('.arrow-indicator');

    instructionsBtn.addEventListener('click', () => {
      instructionsContent.classList.toggle('hidden');
      instructionsArrow.classList.toggle('rotated');
    });
  }

  // ==========================================
  // 4. FILM THEATER FULLSCREEN MODAL
  // ==========================================
  const modal = document.getElementById('video-modal');
  const modalIframe = document.getElementById('modal-iframe');
  const modalCloseBtn = document.getElementById('modal-close-btn');
  const modalCategory = document.getElementById('modal-category');
  const modalTitle = document.getElementById('modal-title');
  const modalDesc = document.getElementById('modal-desc');

  let wasAudioPlayingBeforeVideo = false;
  let savedAudioVolume = 1;

  function pauseBackgroundMusicForVideo() {
    savedAudioVolume = audio.volume;
    wasAudioPlayingBeforeVideo = !audio.paused && !audio.ended;

    if (wasAudioPlayingBeforeVideo) {
      audio.volume = 0.15;
      audio.pause();
    }
  }

  function restoreBackgroundMusicAfterVideo() {
    if (wasAudioPlayingBeforeVideo) {
      audio.volume = savedAudioVolume;
      audio.play().catch(err => {
        console.warn('Background audio could not resume after video close.', err);
      });
    }
  }

  function openVideoModal(videoData) {
    pauseBackgroundMusicForVideo();

    modalCategory.textContent = videoData.category;
    modalTitle.textContent = videoData.title;
    modalDesc.textContent = videoData.description;
    
    // Inject the Google Drive iframe link
    modalIframe.src = `https://drive.google.com/file/d/${videoData.id}/preview`;
    
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden'; // Lock background scroll
  }

  function closeVideoModal() {
    restoreBackgroundMusicAfterVideo();

    // Zero out iframe src to prevent ongoing audio/playback in hidden frame
    modalIframe.src = '';
    
    modal.classList.add('hidden');
    document.body.style.overflow = 'auto'; // Unlock scroll
  }

  modalCloseBtn.addEventListener('click', closeVideoModal);
  
  // Close modal when clicking backdrop area
  document.querySelector('.cinema-backdrop').addEventListener('click', closeVideoModal);

  // Keyboard escape key close
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !modal.classList.contains('hidden')) {
      closeVideoModal();
    }
  });

  // ==========================================
  // 5. TYPEWRITER EFFECT IN HERO SECTION
  // ==========================================
  const typedSubtitleEl = document.getElementById('typed-subtitle');
  if (typedSubtitleEl) {
    const phrases = [
      "A tapestry of late nights, shared laughs, and memories that became family.",
      "Four years of surviving exams, 3 AM chai runs, and finding lifelong friends.",
      "Reliving every sleepy lecture, backseat debate, and stage performance.",
      "From nervous first-day hellos to teary-eyed final goodbyes."
    ];
    let phraseIndex = 0;
    let characterIndex = 0;
    let isDeleting = false;
    let typingSpeed = 50; // Typing speed in ms

    function type() {
      const currentPhrase = phrases[phraseIndex];
      
      if (isDeleting) {
        // Deleting characters
        typedSubtitleEl.textContent = currentPhrase.substring(0, characterIndex - 1);
        characterIndex--;
        typingSpeed = 20; // Faster deleting speed
      } else {
        // Typing characters
        typedSubtitleEl.textContent = currentPhrase.substring(0, characterIndex + 1);
        characterIndex++;
        typingSpeed = 50; // Normal typing speed
      }

      // If full phrase is typed
      if (!isDeleting && characterIndex === currentPhrase.length) {
        isDeleting = true;
        typingSpeed = 2000; // Pause at the end of the phrase
      } 
      // If phrase is fully deleted
      else if (isDeleting && characterIndex === 0) {
        isDeleting = false;
        phraseIndex = (phraseIndex + 1) % phrases.length;
        typingSpeed = 500; // Pause before typing the next phrase
      }

      setTimeout(type, typingSpeed);
    }

    // Start typing animation
    type();
  }

});

