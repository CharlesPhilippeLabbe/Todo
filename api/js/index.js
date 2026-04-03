//const hash = window.location.hash.substring(1)
//if(hash !== ''){

//    const element = document.getElementById("list-"+hash)
//    if(element){
//        element.scrollIntoView({behavior: "smooth"})
//    }

//    const input = document.getElementById("task-input-"+hash)
//    if(input){
//        input.focus()
//    }
//}
function getLists() {
    const url = new URL(window.location.href);
    return url.searchParams.getAll("lists")
}

let state = getLists()
setChecked("default", true)
for (let x of state) {
    setChecked(x, true)
}

function setChecked(x, value) {
    const el = document.getElementById(`select-list-${x}`)
    if (!el) {
        return
    }
    el.checked = value;
}
document.addEventListener('htmx:afterSwap', function() {
})

document.addEventListener('htmx:beforeSwap', function(evt) {
    if (evt.detail.xhr.status === 422) {
        evt.detail.shouldSwap = true;
        evt.detail.isError = false;
    }

});

document.addEventListener('afterRemove', function(e) {
    if (!e?.detail?.value) {
        return
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("lists")
    setChecked(e.detail.value, false)
    state = state.filter(x => {//Work around to make it work with old safari
        if (x == e.detail.value) {
            return false
        }
        url.searchParams.append("lists", x)
        return true
    })
    window.history.pushState({ path: url.href }, '', url.href)
});

document.addEventListener('afterAdd', function(e) {
    if (!e?.detail?.value) {
        return
    }
    setChecked(e.detail.value, true)
    if (e.detail.value == "default") {
        return
    }
    const url = new URL(window.location.href);
    url.searchParams.append("lists", e.detail.value);
    state.push(e.detail.value)
    window.history.pushState({ path: url.href }, '', url.href)
});

document.body.addEventListener('htmx:syntax:error', function(event) {
    console.error('HTMX Syntax Error:', event.detail);
});


function toggleFromDataset(el, f) {
    const linkedId = el?.dataset?.linkedId
    if (!linkedId) {
        return
    }
    const list = document.getElementById(linkedId)
    if (!list) {
        return
    }
    list.classList[f]("active")
}
function onEnter(e) {
    toggleFromDataset(e.target, "add")
}
function onLeave(e) {
    toggleFromDataset(e.target, "remove")
}

