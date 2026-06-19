
    function compressImage(file, maxWidth, maxHeight, quality) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = (event) => {
          const img = new Image();
          img.src = event.target.result;
          img.onload = () => {
            let width = img.width;
            let height = img.height;

            if (width > maxWidth || height > maxHeight) {
              if (width > height) {
                height = Math.round((height * maxWidth) / width);
                width = maxWidth;
              } else {
                width = Math.round((width * maxHeight) / height);
                height = maxHeight;
              }
            }

            const canvas = document.createElement('canvas');
            canvas.width = width;
            canvas.height = height;

            const ctx = canvas.getContext('2d');
            ctx.drawImage(img, 0, 0, width, height);

            canvas.toBlob(
              (blob) => {
                if (blob) {
                  // Keep original file base name but output as optimized jpeg
                  const cleanName = file.name.replace(/\.[^/.]+$/, "") + ".jpg";
                  const compressedFile = new File([blob], cleanName, {
                    type: 'image/jpeg',
                    lastModified: Date.now()
                  });
                  resolve(compressedFile);
                } else {
                  reject(new Error('Canvas to Blob conversion failed'));
                }
              },
              'image/jpeg',
              quality
            );
          };
          img.onerror = (err) => reject(err);
        };
        reader.onerror = (err) => reject(err);
      });
    }

    document.addEventListener('DOMContentLoaded', () => {
      const form = document.getElementById('admin-upload-form');
      const fileInput = document.getElementById('media-file');
      const durationGroup = document.getElementById('duration-group');
      const progressContainer = document.getElementById('progress-container');
      const progressFill = document.getElementById('progress-fill');
      const statusText = document.getElementById('upload-status');
      
      // Toggle duration input based on media type
      document.querySelectorAll('input[name="media-type"]').forEach(radio => {
        radio.addEventListener('change', (e) => {
          if (e.target.value === 'photo') {
            durationGroup.style.display = 'none';
          } else {
            durationGroup.style.display = 'flex';
          }
        });
      });

      // Handle Media Upload
      form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const secret = document.getElementById('admin-secret').value;
        const title = document.getElementById('media-title').value;
        const type = document.querySelector('input[name="media-type"]:checked').value;
        const duration = type === 'video' ? document.getElementById('media-duration').value : '';
        const description = document.getElementById('media-desc').value;
        const file = fileInput.files[0];

        if (!file) return;

        progressContainer.style.display = 'block';
        progressFill.style.width = '0%';

        if (type === 'photo') {
          statusText.textContent = 'Optimizing image for fast loading...';
          // Resize to max 1920x1920 and compress with 80% quality
          compressImage(file, 1920, 1920, 0.8)
            .then(optimizedFile => {
              startUpload(optimizedFile);
            })
            .catch(err => {
              console.warn('Image compression failed, using original:', err);
              startUpload(file);
            });
        } else {
          startUpload(file);
        }

        function startUpload(fileToUpload) {
          statusText.textContent = 'Requesting upload authorization...';

          fetch('/api/graphql', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({
              query: `
                mutation GenerateUploadURL($secret: String!, $filename: String!) {
                  generateUploadURL(secret: $secret, filename: $filename) {
                    uploadUrl
                    publicUrl
                    key
                  }
                }
              `,
              variables: {
                secret: secret,
                filename: fileToUpload.name
              }
            })
          })
          .then(async res => {
            const result = await res.json();
            if (result.errors) throw new Error(result.errors[0].message || 'Failed to authorize upload');
            return result.data.generateUploadURL;
          })
          .then(authData => {
            statusText.textContent = 'Uploading file directly to storage...';
            
            const xhr = new XMLHttpRequest();
            xhr.open('PUT', authData.uploadUrl, true);
            xhr.setRequestHeader('Content-Type', fileToUpload.type);

            xhr.upload.onprogress = (evt) => {
              if (evt.lengthComputable) {
                const percentComplete = Math.round((evt.loaded / evt.total) * 100);
                progressFill.style.width = percentComplete + '%';
                statusText.textContent = `Uploading: ${percentComplete}% (${Math.round(evt.loaded/1024/1024 * 10)/10}MB / ${Math.round(evt.total/1024/1024 * 10)/10}MB)`;
              }
            };

            xhr.onload = () => {
              if (xhr.status === 200) {
                statusText.textContent = 'Saving file metadata to database...';
                
                fetch('/api/graphql', {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json'
                  },
                  body: JSON.stringify({
                    query: `
                      mutation SaveMediaMetadata($secret: String!, $url: String!, $title: String!, $type: String!, $description: String, $duration: String) {
                        saveMediaMetadata(secret: $secret, url: $url, title: $title, type: $type, description: $description, duration: $duration) {
                          id
                        }
                      }
                    `,
                    variables: {
                      secret: secret,
                      url: authData.publicUrl,
                      title: title,
                      type: type,
                      description: description,
                      duration: duration
                    }
                  })
                })
                .then(async res => {
                  const result = await res.json();
                  if (result.errors) throw new Error(result.errors[0].message || 'Failed to save metadata');
                  return result.data.saveMediaMetadata;
                })
                .then(() => {
                  statusText.textContent = 'Upload Complete! Memory is live.';
                  progressFill.style.background = '#6b7a67';
                  alert('Success! Memory is now live in the vault.');
                  form.reset();
                  durationGroup.style.display = 'flex';
                  progressContainer.style.display = 'none';
                })
                .catch(err => {
                  statusText.textContent = 'Metadata Save Failed';
                  alert('Error saving metadata to database: ' + err.message);
                });

              } else {
                statusText.textContent = 'Storage Upload Failed';
                alert('Error uploading file to storage bucket. Status: ' + xhr.status);
              }
            };

            xhr.onerror = () => {
              statusText.textContent = 'Network Error during upload';
              alert('A network error occurred while uploading to storage.');
            };

            xhr.send(fileToUpload);
          })
          .catch(err => {
            progressContainer.style.display = 'none';
            statusText.textContent = 'Authorization Failed';
            alert('Upload authorization failed: ' + err.message);
          });
        }
      });

      // Media list refresh and delete logic
      const refreshMediaBtn = document.getElementById('refresh-media-btn');
      const mediaTbody = document.getElementById('admin-media-tbody');

      function loadMediaList() {
        mediaTbody.innerHTML = `<tr><td colspan="4" style="text-align: center; color: var(--text-muted); padding: 20px 10px;">Loading media assets...</td></tr>`;
        
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
                }
              }
            `
          })
        })
          .then(res => {
            if (!res.ok) throw new Error('Failed to load media list');
            return res.json();
          })
          .then(result => {
            if (result.errors) throw new Error(result.errors[0].message);
            const data = result.data.listMedia;
            mediaTbody.innerHTML = '';
            if (data.length === 0) {
              mediaTbody.innerHTML = `<tr><td colspan="4" style="text-align: center; color: var(--text-muted); padding: 20px 10px;">No media assets found in database.</td></tr>`;
              return;
            }

            data.forEach(item => {
              const tr = document.createElement('tr');
              
              let previewHtml = '';
              if (item.type === 'photo') {
                previewHtml = `<img src="${item.url}" class="media-thumbnail-mini" alt="Preview">`;
              } else {
                previewHtml = `<div style="width: 40px; height: 40px; display: flex; align-items: center; justify-content: center; background: var(--sage-green-light); border-radius: 4px; border: 1px solid var(--border-sepia); font-size: 1.2rem;">🎬</div>`;
              }

              tr.innerHTML = `
                <td>${previewHtml}</td>
                <td style="font-weight: 500;">${item.title}</td>
                <td><span style="text-transform: capitalize;">${item.type}</span></td>
                <td>
                  <button class="btn-delete" data-id="${item.id}">Delete</button>
                </td>
              `;
              
              tr.querySelector('.btn-delete').addEventListener('click', (e) => {
                const mediaId = e.target.getAttribute('data-id');
                const secret = document.getElementById('admin-secret').value;
                if (!secret) {
                  alert('Please enter your Admin Secret Key in the upload form first.');
                  document.getElementById('admin-secret').focus();
                  return;
                }

                if (!confirm(`Are you sure you want to permanently delete the media asset "${item.title}" from the database and Backblaze B2 storage?`)) {
                  return;
                }

                e.target.disabled = true;
                e.target.textContent = 'Deleting...';

                fetch('/api/graphql', {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json'
                  },
                  body: JSON.stringify({
                    query: `
                      mutation DeleteMedia($secret: String!, $id: ID!) {
                        deleteMedia(secret: $secret, id: $id)
                      }
                    `,
                    variables: {
                      secret: secret,
                      id: mediaId
                    }
                  })
                })
                .then(async res => {
                  const result = await res.json();
                  if (result.errors) throw new Error(result.errors[0].message || 'Failed to delete media asset');
                  return result.data;
                })
                .then(() => {
                  alert('Media asset deleted successfully!');
                  loadMediaList();
                })
                .catch(err => {
                  alert('Error deleting media: ' + err.message);
                  e.target.disabled = false;
                  e.target.textContent = 'Delete';
                });
              });

              mediaTbody.appendChild(tr);
            });
          })
          .catch(err => {
            mediaTbody.innerHTML = `<tr><td colspan="5" style="text-align: center; color: #7c3d49; padding: 20px 10px; font-weight: bold;">Error: ${err.message}</td></tr>`;
          });
      refreshMediaBtn.addEventListener('click', loadMediaList);

      // Mobile Navigation Drawer Toggle
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
  