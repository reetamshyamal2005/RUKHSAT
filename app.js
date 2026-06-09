document.addEventListener('DOMContentLoaded', () => {

  // ==========================================
  // 1. AUDIO CONTROLLER
  // ==========================================
  const audio = document.getElementById('bg-audio');
  const quickMusicBtn = document.getElementById('music-quick-btn');
  const quickMusicIcon = quickMusicBtn?.querySelector('.music-icon');
  const quickMusicText = quickMusicBtn?.querySelector('.music-status-text');

  let isPlaying = false;

  function updateMusicUI(playing) {
    isPlaying = playing;
    if (playing) {
      if (quickMusicIcon) quickMusicIcon.classList.add('playing');
      if (quickMusicText) quickMusicText.textContent = 'Pause Soundtrack';
    } else {
      if (quickMusicIcon) quickMusicIcon.classList.remove('playing');
      if (quickMusicText) quickMusicText.textContent = 'Play Soundtrack';
    }
  }

  if (audio) {
    // Sync UI with audio state
    audio.addEventListener('play', () => updateMusicUI(true));
    audio.addEventListener('pause', () => updateMusicUI(false));

    // Handle initial autoplay attempt
    const playPromise = audio.play();
    if (playPromise !== undefined) {
      playPromise.then(() => {
        updateMusicUI(true);
      }).catch(error => {
        // Auto-play was prevented
        updateMusicUI(false);
      });
    }
  }

  function toggleMusic() {
    if (!audio) return;
    if (isPlaying) {
      audio.pause();
    } else {
      audio.play().catch(err => {
        console.warn("Playback prevented", err);
      });
    }
  }

  if (quickMusicBtn) {
    quickMusicBtn.addEventListener('click', toggleMusic);
  }

  // ==========================================
  // 2. HERO TYPING ANIMATION
  // ==========================================
  const heroTypingText = document.getElementById('hero-typing-text');
  if (heroTypingText) {
    const heroMessages = [
      "Goodbyes are not forever; they are simply 'see you later'",
      "A chapter ends, a lifetime of memories remains.",
      "Years will pass, faces may change, but these moments will remain timeless.",
      "From strangers to friends, from friends to family - this is our story.",
    ];
    let messageIndex = 0;
    let currentCharIndex = 0;
    let isDeleting = false;

    function animateHeroText() {
      const currentMessage = heroMessages[messageIndex];
      if (!isDeleting) {
        heroTypingText.textContent = currentMessage.slice(0, currentCharIndex + 1);
        currentCharIndex += 1;
        if (currentCharIndex === currentMessage.length) {
          isDeleting = true;
          setTimeout(animateHeroText, 1800);
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
      setTimeout(animateHeroText, isDeleting ? 30 : 45);
    }
    animateHeroText();
  }

  // Sticky Navbar Shrink Effect
  window.addEventListener('scroll', () => {
    const navbar = document.querySelector('.navbar');
    if (navbar) {
      if (window.scrollY > 50) navbar.classList.add('scrolled');
      else navbar.classList.remove('scrolled');
    }
  });

  // ==========================================
  // 3. VIDEO LOUNGE RENDERING & FILTERS
  // ==========================================
  const gridContainer = document.getElementById('video-grid-container');
  const photoGridContainer = document.getElementById('photo-grid-container');
  const searchInput = document.getElementById('video-search');
  const filterBtns = document.querySelectorAll('.filter-btn');
  
  const CATEGORY_ILLUSTRATIONS = {
    campus: `<svg viewBox="0 0 100 100" class="retro-doodle"><rect x="10" y="30" width="80" height="50" rx="3" fill="#6b7a67" opacity="0.3"/><polygon points="50,10 5,30 95,30" fill="#7c3d49" opacity="0.8"/><line x1="50" y1="30" x2="50" y2="80" stroke="#fcfbf7" stroke-width="2"/><rect x="25" y="45" width="12" height="12" rx="1" fill="#fff" opacity="0.7"/><rect x="63" y="45" width="12" height="12" rx="1" fill="#fff" opacity="0.7"/><path d="M44,80 L44,68 L56,68 L56,80 Z" fill="#3a2518"/></svg>`,
    classroom: `<svg viewBox="0 0 100 100" class="retro-doodle"><rect x="15" y="20" width="70" height="42" fill="#2d2926" stroke="#c2945d" stroke-width="3"/><line x1="25" y1="28" x2="75" y2="28" stroke="#c2945d" stroke-width="1.5" stroke-dasharray="3"/><text x="32" y="44" fill="#c2945d" font-family="Caveat" font-size="14" font-weight="bold">1+1 = Memories</text><path d="M40,75 L15,92 M60,75 L85,92" stroke="#2d2926" stroke-width="3" stroke-linecap="round"/><rect x="35" y="62" width="30" height="15" fill="#c2945d" rx="2"/></svg>`,
    fests: `<svg viewBox="0 0 100 100" class="retro-doodle"><circle cx="50" cy="50" r="38" fill="#7c3d49" opacity="0.2"/><path d="M25,65 Q35,30 50,65 T75,65" fill="none" stroke="#7c3d49" stroke-width="4" stroke-linecap="round"/><circle cx="25" cy="65" r="5" fill="#c2945d"/><circle cx="50" cy="65" r="5" fill="#c2945d"/><circle cx="75" cy="65" r="5" fill="#c2945d"/><path d="M20,30 L30,40 L40,30 L50,40 L60,30 L70,40 L80,30" fill="none" stroke="#6b7a67" stroke-width="3"/></svg>`,
    hostel: `<svg viewBox="0 0 100 100" class="retro-doodle"><rect x="20" y="25" width="60" height="50" rx="4" fill="none" stroke="#2d2926" stroke-width="3"/><line x1="20" y1="50" x2="80" y2="50" stroke="#2d2926" stroke-width="3"/><path d="M30,35 H70 M30,60 H70" stroke="#6b7a67" stroke-width="2" stroke-linecap="round"/><circle cx="35" cy="42" r="3" fill="#7c3d49"/><circle cx="65" cy="67" r="3" fill="#c2945d"/></svg>`,
    messages: `<svg viewBox="0 0 100 100" class="retro-doodle"><path d="M10,25 L90,25 L50,55 Z" fill="#f4efdf" stroke="#2d2926" stroke-width="3" stroke-linejoin="round"/><rect x="10" y="25" width="80" height="50" fill="none" stroke="#2d2926" stroke-width="3" stroke-linejoin="round"/><path d="M10,75 L42,48 M90,75 L58,48" stroke="#2d2926" stroke-width="3"/><path d="M50,40 Q40,30 50,20 Q60,30 50,40 Z" fill="#7c3d49"/></svg>`
  };

  function renderVideos(videosArray) {
    if (!gridContainer) return; 
    gridContainer.innerHTML = '';
    
    if (videosArray.length === 0) {
      gridContainer.innerHTML = `<div class="no-results-message"><p class="scribble" style="font-size: 2.2rem; text-align: center; grid-column: 1 / -1; width: 100%; color: var(--rose-burgundy);">"No matching reels found..."</p></div>`;
      return;
    }

    videosArray.forEach((video, index) => {
      const card = document.createElement('div');
      card.className = 'polaroid-card';
      
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
          <div class="play-overlay"><div class="play-icon-circle"><svg viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg></div></div>
          <span class="video-length-badge">${video.duration}</span>
        </div>
        <div class="polaroid-body">
          <span class="polaroid-cat-tag">${video.category}</span>
          <h4 class="polaroid-card-title">${video.title}</h4>
          <p class="polaroid-card-desc">${video.description}</p>
        </div>
      `;
      card.addEventListener('click', () => openVideoModal(video));
      gridContainer.appendChild(card);
    });
  }

  function renderPhotos(photosArray) {
    if (!photoGridContainer) return;
    photoGridContainer.innerHTML = '';

    if (photosArray.length === 0) {
      photoGridContainer.innerHTML = `<div class="no-results-message"><p class="scribble" style="font-size: 2.2rem; text-align: center; grid-column: 1 / -1; width: 100%; color: var(--rose-burgundy);">"No captured photos found yet..."</p></div>`;
      return;
    }

    photosArray.forEach((photo, index) => {
      const card = document.createElement('div');
      card.className = 'polaroid-card';
      
      card.innerHTML = `
        <div class="tape-sticker" style="top: -14px; left: 50%; transform: translateX(-50%) rotate(${index % 2 === 0 ? '-2' : '2'}deg); width: 85px; height: 26px;"></div>
        <div class="video-thumbnail-container">
          <img src="${photo.url}" alt="${photo.title}" class="video-thumbnail" style="filter: sepia(0.15) contrast(1.05); object-fit: cover;" loading="lazy">
          <div class="play-overlay" style="opacity: 0; background: rgba(124, 61, 73, 0.4);"><span class="scribble" style="color: white; font-size: 1.5rem;">View Photo</span></div>
        </div>
        <div class="polaroid-body">
          <span class="polaroid-cat-tag">${photo.category}</span>
          <h4 class="polaroid-card-title">${photo.title}</h4>
          <p class="polaroid-card-desc">${photo.description}</p>
        </div>
      `;
      card.addEventListener('click', () => openPhotoModal(photo));
      photoGridContainer.appendChild(card);
    });
  }

  let activeMediaDatabase = [];

  function loadMedia() {
    if (!gridContainer && !photoGridContainer) return;

    fetch('/api/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        query: `
          query ListMedia {
            listMedia {
              id
              url
              title
              type
              category
              description
              duration
              createdAt
            }
          }
        `
      })
    })
      .then(res => {
        if (!res.ok) throw new Error('API failed');
        return res.json();
      })
      .then(result => {
        if (result.errors) {
          throw new Error(result.errors[0].message);
        }
        const data = result.data.listMedia;
        if (gridContainer) {
          const videosOnly = data.filter(item => item.type === 'video');
          if (videosOnly.length > 0) {
            activeMediaDatabase = videosOnly;
          } else if (typeof VIDEO_DATABASE !== 'undefined') {
            activeMediaDatabase = VIDEO_DATABASE;
          } else {
            activeMediaDatabase = [];
          }
          renderVideos(activeMediaDatabase);
        } else if (photoGridContainer) {
          const photosOnly = data.filter(item => item.type === 'photo');
          renderPhotos(photosOnly);
        }
      })
      .catch(err => {
        console.warn('API error, falling back to static database:', err);
        if (gridContainer) {
          if (typeof VIDEO_DATABASE !== 'undefined') {
            activeMediaDatabase = VIDEO_DATABASE;
          } else {
            activeMediaDatabase = [];
          }
          renderVideos(activeMediaDatabase);
        } else if (photoGridContainer) {
          renderPhotos([]);
        }
      });
  }

  // Load initial media
  if (gridContainer || photoGridContainer) {
    loadMedia();
  }

  function handleFilterChange() {
    if (!searchInput) return;
    const searchQuery = searchInput.value.toLowerCase();
    const activeCategoryBtn = document.querySelector('.filter-btn.active');
    if (!activeCategoryBtn) return;
    const selectedCategory = activeCategoryBtn.getAttribute('data-category');

    const filteredList = activeMediaDatabase.filter(video => {
      const matchesSearch = video.title.toLowerCase().includes(searchQuery) || video.description.toLowerCase().includes(searchQuery);
      const matchesCategory = selectedCategory === 'all' || video.category === selectedCategory;
      return matchesSearch && matchesCategory;
    });
    renderVideos(filteredList);
  }

  if (searchInput) searchInput.addEventListener('input', handleFilterChange);

  if (filterBtns) {
    filterBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        filterBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        handleFilterChange();
      });
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
    if (!audio) return;
    savedAudioVolume = audio.volume;
    wasAudioPlayingBeforeVideo = !audio.paused && !audio.ended;
    if (wasAudioPlayingBeforeVideo) {
      audio.volume = 0.15;
      audio.pause();
    }
  }

  function restoreBackgroundMusicAfterVideo() {
    if (!audio) return;
    if (wasAudioPlayingBeforeVideo) {
      audio.volume = savedAudioVolume;
      audio.play().catch(err => console.warn('Audio error', err));
    }
  }

  function openVideoModal(videoData) {
    if (!modal) return;
    pauseBackgroundMusicForVideo();
    if (modalCategory) modalCategory.textContent = videoData.category;
    if (modalTitle) modalTitle.textContent = videoData.title;
    if (modalDesc) modalDesc.textContent = videoData.description;
    if (modalIframe) modalIframe.src = videoData.url;
    
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden'; 
  }

  function closeVideoModal() {
    if (!modal) return;
    restoreBackgroundMusicAfterVideo();
    if (modalIframe) modalIframe.src = '';
    modal.classList.add('hidden');
    document.body.style.overflow = 'auto'; 
  }

  if (modalCloseBtn) modalCloseBtn.addEventListener('click', closeVideoModal);
  
  const backdrop = document.querySelector('.cinema-backdrop');
  if (backdrop) backdrop.addEventListener('click', closeVideoModal);

  // Photo modal elements and events
  const photoModal = document.getElementById('photo-modal');
  const photoModalImg = document.getElementById('photo-modal-img');
  const photoModalClose = document.getElementById('photo-modal-close-btn');
  const photoModalCat = document.getElementById('photo-modal-category');
  const photoModalTitle = document.getElementById('photo-modal-title');
  const photoModalDesc = document.getElementById('photo-modal-desc');

  function openPhotoModal(photo) {
    if (!photoModal) return;
    if (photoModalCat) photoModalCat.textContent = photo.category;
    if (photoModalTitle) photoModalTitle.textContent = photo.title;
    if (photoModalDesc) photoModalDesc.textContent = photo.description;
    if (photoModalImg) photoModalImg.src = photo.url;
    photoModal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  }

  function closePhotoModal() {
    if (!photoModal) return;
    photoModal.classList.add('hidden');
    document.body.style.overflow = 'auto';
  }

  if (photoModalClose) photoModalClose.addEventListener('click', closePhotoModal);
  const photoBackdrop = photoModal?.querySelector('.cinema-backdrop');
  if (photoBackdrop) photoBackdrop.addEventListener('click', closePhotoModal);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      if (modal && !modal.classList.contains('hidden')) {
        closeVideoModal();
      }
      if (photoModal && !photoModal.classList.contains('hidden')) {
        closePhotoModal();
      }
    }
  });

  // ==========================================
  // 5. MOBILE NAVIGATION DRAWER
  // ==========================================
  const menuToggle = document.getElementById('menu-toggle');
  const menuClose = document.getElementById('menu-close');
  const navDrawer = document.getElementById('nav-drawer');

  function openDrawer() {
    if (navDrawer) {
      navDrawer.classList.add('open');
      document.body.style.overflow = 'hidden'; 
    }
  }

  function closeDrawer() {
    if (navDrawer) {
      navDrawer.classList.remove('open');
      document.body.style.overflow = 'auto'; 
    }
  }

  if (menuToggle) menuToggle.addEventListener('click', openDrawer);
  if (menuClose) menuClose.addEventListener('click', closeDrawer);

  const drawerLinks = navDrawer?.querySelectorAll('a');
  if (drawerLinks) {
    drawerLinks.forEach(link => {
      link.addEventListener('click', closeDrawer);
    });
  }
});