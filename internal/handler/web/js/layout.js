function loadCSS(url, callback, errorCallback, attributes = {}) {
    const existingLinks = document.querySelectorAll('link[rel="stylesheet"]');
    for (let i = 0; i < existingLinks.length; i++) {
        if (existingLinks[i].href === url || existingLinks[i].href.endsWith(url)) {
            if (typeof callback === 'function') {
                callback();
            }
            return existingLinks[i];
        }
    }

    const link = document.createElement('link');

    link.rel = attributes.rel || 'stylesheet';
    link.type = attributes.type || 'text/css';
    link.href = url;

    if (attributes.media) link.media = attributes.media;
    if (attributes.id) link.id = attributes.id;
    if (attributes.crossOrigin) link.crossOrigin = attributes.crossOrigin;

    link.onload = function () {
        if (typeof callback === 'function') {
            callback();
        }
    };

    link.onerror = function (err) {
        if (typeof errorCallback === 'function') {
            errorCallback(err);
        } else {
            console.error(`Failed to load CSS: ${url}`, err);
        }
    };

    const head = document.head || document.getElementsByTagName('head')[0];
    head.appendChild(link);

    return link;
}

function loadCSSAsPromise(url, attributes = {}) {
    return new Promise((resolve, reject) => {
        loadCSS(
            url,
            resolve,
            reject,
            attributes
        );
    });
}

function loadMD5Script() {
    return new Promise((resolve, reject) => {
        if (typeof md5 !== 'undefined') {
            resolve(); // MD5 library already loaded
            return;
        }
        const script = document.createElement('script');
        script.src = '/js/md5.min.js';
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => {
            console.error('Failed to load MD5 library');
            reject(new Error('MD5 library load failed'));
        };
        document.head.appendChild(script);
    });
}

function createPage() {
    const main = document.createElement('div');
    main.id = "page";
    document.body.appendChild(main);
}