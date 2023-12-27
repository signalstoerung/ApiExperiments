const breakingNewsSection = document.getElementById("breaking");
const articleTemplate = document.getElementById("articleTemplate");

if (!("content" in document.createElement("template"))) {
    console.error("<template> not supported in this browser")
} else {
    loadNews();
//    appendArticle(breakingNewsSection, "Howdy", "https://ph6point6.com/")
}

function loadNews(){
    const apiUrl = "/breaking/";
    fetch(apiUrl).then( (response) => {
       if (response.ok) {
        return response.json();
       } else {
        console.error(response.statusText);
       } 
    }).then( (json) => {
        processUpdates(json);
    }).catch( (e) => {
        console.error("Something went wrong: ",e);
    })
}

function processUpdates(json) {
    if (!Array.isArray(json)) {
        console.error("Expected Array, got something else!");
        console.log(json);
        return;
    }
    for (const story of json) {
        const updates = story["Updates"];
        const latest = updates[updates.length-1];
        appendArticle(breakingNewsSection, latest["Headline"], latest["Url"]);
    }
}

function appendArticle(node, headlineText, url) {
    let article = articleTemplate.content.cloneNode(true);
    const headline = article.querySelector("#headline");
    headline.innerText = headlineText;
    const link = article.querySelector("#link");
    link.innerHTML = `<a href="${url}" target="_blank">Link</a>`;
    node.appendChild(article);
}