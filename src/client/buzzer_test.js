const regex = /(^|[^@]*\W)(@\w+)/g
const msg = "Hi @tom & @jerry!"

function linkify(msg) {
    const children = []
    const matches = msg.match(regex) || []

    let lastIndex = 0
    matches.forEach(match => {
        let results = /(^|\W)(@\w+)/.exec(match)
        const start = lastIndex + results.index + 1
        if (lastIndex > 0 || start > 1) {
            children.push(msg.slice(lastIndex, start))
        }
        lastIndex += results.input.length
        children.push(`<a>${results[2]}</a>`)
        // console.error(results)
    })
    children.push(msg.slice(lastIndex))
    return children
}
console.log(linkify("Hi @tom & @jerry!").join(''))
console.log(linkify("Yo yo yo").join(''))
console.log(linkify("@taeber says").join(''))
console.log(linkify("@taeber").join(''))
console.log(linkify("This should not work email@taeber.com").join(''))
console.log(linkify("where the party @").join(''))