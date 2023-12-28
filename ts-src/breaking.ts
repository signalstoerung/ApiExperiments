const breakingNewsSection = document.getElementById("breaking");
const articleTemplate = document.getElementById("articleTemplate") as HTMLTemplateElement;
const refreshButton = document.getElementById("refresh");
refreshButton.addEventListener("click", refresh);
const notifyButton = document.getElementById("notify");
notifyButton.addEventListener("click", activateNotifications);
const messageField = document.getElementById("message");
let notificationsOn: boolean;
let notificationTimer: number = -1;
let updateTimer: number = -1;
const notificationCheckInterval = 60000 // 60 seconds
const refreshFromBackendInterval = 300000 // 5 minutes

type Update = {
    ID: number,
    CreatedAt: string,
    Headline: string,
    Body: string,
    BodyHash: string,
    Url: string,
    DevelopingStoryID: number,
    New: boolean,
}

type Story = {
    ID: number,
    CreatedAt: string,
    Slug: string,
    Updates: Update[],
    LastUpdated: string,
}

class StoriesUpdates {
    allUpdates: Update[]
    constructor(){
        this.allUpdates = [];
    }
    newStories(): Update[]{
        return this.allUpdates.filter( (update) => update.New )
    }
    clearNew(){
        for (let update of this.allUpdates){
            update.New=false;
        }
    }
    processUpdate(obj: Story[]){
        // previously, we were clearing all the "new" flags here - but this is now done by the "notify" function.
        // New==true == not notified yet
        // for (let i=0; i<this.allUpdates.length; i++){
        //     this.allUpdates[i].New=false;
        // }
        // process new updates
        for (const story of obj) {
            for (const update of story.Updates){
                const found = this.allUpdates.findIndex((value)=>{
                    return value.BodyHash==update.BodyHash;
                })
                if (found != -1) {
                    // have this update already, skip
                    continue
                }
                update.New = true;
                this.allUpdates.push(update);
            }
        }
    }
}

const myUpdates = new StoriesUpdates();

if (!("content" in document.createElement("template"))) {
    console.error("<template> not supported in this browser")
} else {
    refresh();
    if (updateTimer == -1) {
        updateTimer = setInterval(refresh,refreshFromBackendInterval);
        messageField.innerText="Automatic updates activated.";
    }
}


function refresh(){
    const apiUrl = "/breaking/";
    fetch(apiUrl).then( (response) => {
       if (response.ok) {
        return response.json();
       } else {
        console.error(response.statusText);
       } 
    }).then( (obj) => {
        myUpdates.processUpdate(obj);
        showUpdates(false);
    }).catch( (e) => {
        console.error("Something went wrong: ",e);
        
    })
}

function showUpdates(newOnly: boolean) {
    // remove old updates
    while (breakingNewsSection.firstChild) {
        breakingNewsSection.removeChild(breakingNewsSection.firstChild);
    }
    let updates: Update[];
    if (newOnly) {
        updates = myUpdates.newStories().sort( (a,b): number => { 
            return b.ID - a.ID
        })
    } else {
        updates = myUpdates.allUpdates.sort( (a,b): number => { 
            return b.ID - a.ID
        })
    }
    for (const update of updates) {
        const updateTime = new Date(update.CreatedAt);
        appendArticle(breakingNewsSection, update["Headline"], updateTime.toLocaleTimeString(), update["Url"], update.New);
    }
}

function appendArticle(node: HTMLElement, headlineText: string, bodyText: string, url: string, newUpdate: boolean) {
    let article = articleTemplate.content.cloneNode(true) as HTMLElement;
    const r = article.querySelector("article");
    if (newUpdate) {
        r.classList.add("newUpdate");
    }
    const headline = article.querySelector("#headline") as HTMLElement;
    headline.innerText = headlineText;
    const text = article.querySelector("#text") as HTMLElement;
    text.innerText = bodyText;
    const link = article.querySelector("#link");
    link.innerHTML = `<a href="${url}" target="_blank">Link</a>`;
    node.appendChild(article);
}

function activateNotifications(){
    if ("Notification" in window) {
        if (Notification.permission == "denied"){
            messageField.innerHTML = "<strong>You have (previously) denied permission to send notifications.</strong>"
            notifyButton.setAttribute("disabled","true");
        }
        if (Notification.permission != "granted") {
            Notification.requestPermission().then( (permission) => {
                if (notificationTimer == -1) {
                    activate();
                }
            })
        } else {
            // permission previously granted, no need to request again
            if (notificationTimer == -1) {
                activate();
            }
        }
    }
}

function activate(){
    // first, we're clearing all the "new" flags. All the latest updates are on the screen; notifications should only be sent for updates that come in new.
    myUpdates.clearNew();
    notificationTimer = setInterval(notify, notificationCheckInterval)
    console.log("Notification timer set: ", notificationTimer)
    messageField.innerText="Notifications activated.";
    notifyButton.setAttribute("disabled","true");
}

function notify(){
    const img = "/android-chrome-192x192.png";
    let notificationDelay = 0;
    for (let i=0; i<myUpdates.allUpdates.length; i++){
        if (myUpdates.allUpdates[i].New) {
            // spacing out multiple notifications
            setTimeout(()=>{
                new Notification("BREAKING:", {body: myUpdates.allUpdates[i].Headline, icon: img})
            },notificationDelay);
            notificationDelay += 5000;
            // remove the "New" flag
            myUpdates.allUpdates[i].New=false;
        }
    }
}