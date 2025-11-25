function welcome() {
    document.title = "Welcome - MeowBox";
    const page = document.getElementById("page");
    let currentIndex = 0;
    const welcomeDiv = document.createElement("div");
    welcomeDiv.classList.add("welcome");
    const welcomeButton = document.createElement("button");
    welcomeButton.classList.add("welcome-button");
    welcomeButton.innerHTML = 'Next <i class="fa-solid fa-circle-arrow-right"></i>';
    welcomeDiv.appendChild(welcomeButton);
    function displayNextLanguage() {
        if (currentIndex >= languages.length) {
            currentIndex = 0;
        }
        const welcomeTitle = document.createElement("h1");
        welcomeTitle.classList.add("welcome-animation")
        welcomeTitle.textContent = languages[currentIndex]["hello"];
        welcomeDiv.appendChild(welcomeTitle);
        setTimeout(() => {
            welcomeDiv.removeChild(welcomeTitle);
            currentIndex++;
            displayNextLanguage();
        }, 3000);
    }
    displayNextLanguage();
    page.appendChild(welcomeDiv);
}

function install() {
}

// Load external script
function loadScript(scriptUrl) {
    return new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = scriptUrl;
        script.async = true;

        script.onload = () => {
            resolve();
            script.remove();
        };

        script.onerror = () => {
            const errorMsg = `Failed to load script: ${scriptUrl}`;
            console.error(`[ScriptLoader] ${errorMsg}`);
            reject(new Error(errorMsg));
        };

        document.body.appendChild(script);
    });
}

// Initialize app
async function initializeApp() {
    try {
        await Promise.all([
            loadScript("/js/layout.js")
        ]);
        createPage();
        await Promise.all([
            loadScript("/js/i18n.js"),
            loadCSSAsPromise("/css/layout.css"),
            loadCSSAsPromise("/css/install.css"),
            loadCSSAsPromise("/font-awesome/css/all.min.css")
        ])
        welcome();
    } catch (error) {
        console.error('App initialization error:', error);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});