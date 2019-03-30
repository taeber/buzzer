const regex = /(^|[^@]*\W)(@\w+)/g
const msg = "Hi @tom & @jerry!"

function linkify2(msg) {
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

const modes = {
    READY: Symbol(''),
    MENTION: Symbol('@'),
    TAG: Symbol('#'),
    WITHIN: Symbol('w')
}

const isWordChar = ch => /\w/.test(ch)

function linkify(msg) {
  msg += '\0'

    const children = []
    let text = ""
    let mode = modes.READY

    for (let i = 0; i < msg.length; i++) {
        const ch = msg[i]
        switch (mode) {
            case modes.READY:
                if (ch === '@' || ch === '#') {
                    if (text.length > 0) {
                        children.push(text)
                        text = ""
                    }
                    mode = ch === '@' ? modes.MENTION : modes.TAG
                } else if (isWordChar(ch)) {
                    mode = modes.WITHIN
                }
                break

            case modes.MENTION:
                if (isWordChar(ch))
                    break

                if (text.length > 0) {
                    children.push(`<a href="#mention">${text}</a>`)
                    text = ""
                }

                if (ch === '@' || ch === '#')
                    mode = modes.WITHIN
                else
                    mode = modes.READY
                break

            case modes.TAG:
                if (isWordChar(ch))
                    break

                if (text.length > 0) {
                    children.push(`<a href="#tag">${text}</a>`)
                    text = ""
                }

                if (ch === '@' || ch === '#')
                    mode = modes.WITHIN
                else
                    mode = modes.READY
                break

            case modes.WITHIN:
                if (!isWordChar(ch))
                    mode = modes.READY
                break
        }
        text += ch
    }

    if (text.length > 0)
      children.push(text)

    return children
}


console.log(linkify("Hi @tom & @jerry!").join(''))
console.log(linkify("Yo yo yo").join(''))
console.log(linkify("@taeber says").join(''))
console.log(linkify("@taeber #knows #best#of@all").join(''))
console.log(linkify("This should not work email@taeber.com").join(''))
console.log(linkify("where the party @").join(''))
console.log(linkify("Howdy @tom").join(''))
console.log(linkify("").join(''))
