if (false) {
    // This is just a trick to get Intellisense working in VS Code.
    let React = require("react"),
        ReactDOM = require("react-dom"),
        moment = require("moment")
}

class BuzzerWebView extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            client: null,
            loggedIn: false,
            loginFormDisabled: false,
            showRegistration: true,
            username: "",
            password: "",
            messages: [],
            status: "",
            compressed: false,
            profile: null,
            following: [],
            topic: null,
        }

        this.getClient = this.getClient.bind(this)
        this.getMessages = this.getMessages.bind(this)
        this.search = this.search.bind(this)

        this.handleLogin = this.handleLogin.bind(this)
        this.handleLogout = this.handleLogout.bind(this)
        this.handleMessage = this.handleMessage.bind(this)
        this.handleMessageClick = this.handleMessageClick.bind(this)
        this.handlePost = this.handlePost.bind(this)
        this.handleRegister = this.handleRegister.bind(this)
        this.handleSearch = this.handleSearch.bind(this)
        this.handleSubscriptionChange = this.handleSubscriptionChange.bind(this)
        this.handleToggleForms = this.handleToggleForms.bind(this)
    }

    componentDidMount() {
        const banner = document.getElementById("banner")
        banner.addEventListener("click", () => {
            this.setState({ compressed: !this.state.compressed }, () => {
                if (!this.state.compressed)
                    window.scrollTo(0, 0)
            })
        })
    }

    render() {
        const {
            handleLogin, handleLogout, handlePost, handleRegister,
            handleSearch, handleSubscriptionChange, handleMessageClick,
            handleToggleForms,
        } = this

        const {
            loggedIn, loginFormDisabled, username, password, messages,
            showRegistration, status, compressed, profile, following, topic
        } = this.state

        if (showRegistration) {
            return (
                <RegistrationForm
                    username={username}
                    password={password}
                    disabled={loginFormDisabled}
                    onSubmit={handleRegister}
                    onCancel={handleToggleForms}
                />
            )
        }

        if (!loggedIn) {
            return (
                <LoginForm
                    username={username}
                    password={password}
                    disabled={loginFormDisabled}
                    onSubmit={handleLogin}
                    onCancel={handleToggleForms}
                />
            )
        }

        let msgs = messages
        if (profile) {
            msgs = msgs.filter(msg =>
                msg.poster.username === profile ||
                (msg.mentions || []).includes(profile)
            )
        } else if (topic) {
            msgs = msgs.filter(msg => (msg.tags || []).includes(topic))
        } else {
            msgs = msgs.filter(msg =>
                msg.poster.username === username ||
                (msg.mentions || []).includes(username) ||
                following.indexOf(msg.poster.username) >= 0
            )
        }

        const messageList = (
            <ul key="messages" className="messages">
                {msgs.map(msg => (
                    <li className="message" key={msg.id}>
                        <div className="poster">
                            <a href="#mention" onClick={handleMessageClick}>@{msg.poster.username}</a>
                            <span className="posted">
                                {moment(msg.posted).fromNow()}
                            </span>
                        </div>
                        <p className="text">{linkify(msg.text, handleMessageClick)}</p>
                    </li>
                ))}
            </ul>
        )

        if (compressed)
            return messageList

        let hero
        if (!profile && !topic) {
            hero = (
                <div key="hero" className="hero">
                    <span className="big">@{username}</span>
                    <button onClick={handleLogout}>Log out</button>
                </div>
            )
        } else if (profile && !topic) {
            hero = (
                <div key="hero" className="hero">
                    <span className="big">@{profile}</span>
                    <button onClick={
                        (e) => { e.preventDefault(); this.setState({ profile: null, topic: null }) }
                    }>
                        Home
                    </button>
                    <button onClick={handleSubscriptionChange}>
                        {following.indexOf(profile) < 0 ? "Subscribe" : "Unsubscribe"}
                    </button>
                </div >
            )
        } else {
            hero = (
                <div key="hero" className="hero">
                    <span className="big">#{topic}</span>
                    <button onClick={
                        (e) => { e.preventDefault(); this.setState({ profile: null, topic: null }) }
                    }>
                        Home
                    </button>
                </div>
            )
        }

        const post = !profile && !topic ? (
            <form key="post" className="PostForm" onSubmit={handlePost}>
                <textarea
                    name="status"
                    placeholder="What do you wanna say?"
                    value={status}
                    onChange={e => this.setState({ status: e.target.value })}
                />
                <button
                    name="post"
                    type="submit"
                    disabled={status === "" || loginFormDisabled}
                >
                    Post
                </button>
            </form>
        ) : null

        const search = !profile && !topic ? (
            <form key="search" className="SearchForm" onSubmit={handleSearch}>
                <input name="query" placeholder="#tag or @username" />
                <button type="submit" name="post" disabled={loginFormDisabled}>
                    Search
                </button>
            </form>
        ) : null

        return [hero, post, search, msgs.length > 0 && messageList]
    }

    async getClient() {
        let { client } = this.state
        if (!client) {
            client = await makeBuzzerClient(this.props.server, this.handleMessage)
            this.setState({ client })
        }
        return client
    }

    async getMessages(username) {
        const client = await this.getClient()
        client.Messages(username)
    }

    /**
     * @param {Event} event
     */
    async handleLogin(event) {
        event.preventDefault()

        this.setState({ loginFormDisabled: true })

        const client = await this.getClient()

        const creds = {
            username: document.getElementsByName("username")[0].value,
            password: document.getElementsByName("password")[0].value,
        }

        try {
            await client.Login(creds.username, creds.password)

            this.setState({
                username: creds.username,
                password: creds.password,
                loginFormDisabled: false,
                loggedIn: true,
                showRegistration: false,
            }, () => this.getMessages(this.state.username))

        } catch (err) {
            console.error(err)
            this.setState({ loginFormDisabled: false })
            alert(`Error! ${err}`)
        }
    }

    /**
     * @param {Event} event
     */
    handleLogout(event) {
        event.preventDefault()

        if (this.state.client && this.state.client.ws)
            this.state.client.ws.close()

        this.setState({
            loggedIn: false,
            password: "",
            client: null,
        })
    }

    handleMessage(msg) {
        const starts = prefix =>
            prefix === msg.slice(0, prefix.length) ? prefix.length : 0

        let at
        if (at = starts("buzz ")) {
            const buzz = JSON.parse(msg.slice(at))
            if (!this.state.messages.some(m => m.id === buzz.id)) {
                this.setState({
                    messages: this.state.messages.concat([buzz])
                })
            }
            return
        }

        if (at = starts("follow ")) {
            const username = msg.slice(at)
            const { following } = this.state
            if (following.indexOf(username) < 0) {
                this.setState({
                    following: following.concat([username])
                })
            }
            return
        }

        if (at = starts("unfollow ")) {
            const username = msg.slice(at)
            const { following } = this.state
            const index = following.indexOf(username)
            if (index >= 0) {
                this.setState({
                    following:
                        following.slice(0, index)
                            .concat(following.slice(index + 1))
                })
            }
            return
        }
    }

    handleMessageClick(event) {
        event.preventDefault()
        this.search(event.target.innerText)
    }

    /**
     * @param {Event} event
     */
    async handlePost(event) {
        event.preventDefault()

        this.setState({ loginFormDisabled: true })

        const client = await this.getClient()

        const message = document.getElementsByName("status")[0].value

        try {
            await client.Post(message)
            this.setState({ status: "" })
        } catch (err) {
            console.error(err)
            alert(`Error! ${err}`)
        } finally {
            this.setState({ loginFormDisabled: false })
        }
    }

    /**
     * @param {Event} event
     */
    async handleRegister(event) {
        event.preventDefault()

        this.setState({ loginFormDisabled: true })

        const client = await this.getClient()

        const creds = {
            username: document.getElementsByName("username")[0].value,
            password: document.getElementsByName("password")[0].value,
        }

        try {
            await client.Register(creds.username, creds.password)
            await client.Login(creds.username, creds.password)

            this.setState({
                loggedIn: true,
                username: creds.username,
                password: creds.password,
                loginFormDisabled: false,
                showRegistration: false,
            }, () => this.getMessages(this.state.username))
        } catch (err) {
            console.error(err)
            this.setState({ loginFormDisabled: false })
            alert(`Error! ${err}`)
        }
    }

    /**
     * @param {Event} event
     */
    handleSearch(event) {
        event.preventDefault()
        const query = document.getElementsByName("query")[0].value

        this.search(query)
    }

    /**
     * @param {Event} event
     */
    async handleSubscriptionChange(event) {
        event.preventDefault()

        const client = await this.getClient()
        const followee = this.state.profile

        try {
            if (this.state.following.indexOf(followee) < 0)
                client.Follow(followee)
            else
                client.Unfollow(followee)
        } catch (err) {
            console.error(err)
            alert(`Error! ${err}`)
        }

        return

    }

    /**
     * @param {Event} event
     */
    handleToggleForms(event) {
        event.preventDefault()
        this.setState({
            showRegistration: !this.state.showRegistration,
            username: document.getElementsByName("username")[0].value,
            password: document.getElementsByName("password")[0].value,
        })
    }

    async search(query) {
        if (!query)
            return

        if (query[0] !== '#' && query[0] !== '@')
            return setTimeout(() => alert("Error! Query must start with # or @."), 1)

        if (query.length < 2)
            return setTimeout(() => alert("Error! Query too short."), 1)


        if (query[0] === '#') {
            this.setState({ loginFormDisabled: true })

            const client = await this.getClient()
            const topic = query.slice(1)

            try {
                client.Tagged(topic)
                this.setState({ topic })
            } catch (err) {
                console.error(err)
                alert(`Error! ${err}`)
            } finally {
                this.setState({ loginFormDisabled: false })
            }

        } else {
            const username = query.slice(1)
            this.getMessages(username)
            this.setState({ profile: username })
        }
    }
}

/// Client

function makeBuzzerClient(server, msgHandler) {
    return new Promise((resolve) => {
        const ws = new WebSocket(server)
        const client = {
            ws,
            Register: register.bind(null, ws),
            Login: login.bind(null, ws),
            Post: post.bind(null, ws),
            Messages: getMessages.bind(null, ws),
            Tagged: tagged.bind(null, ws),
            Follow: follow.bind(null, ws),
            Unfollow: unfollow.bind(null, ws),
        }

        client.ws.addEventListener("close", () => {
            console.log("BuzzerClient: closed")
            alert("Lost connection to server")
            // window.location = window.location
        })
        client.ws.addEventListener("message", (e) => {
            console.log("BuzzerClient: recv:", e.data)
            msgHandler(e.data)
        })
        client.ws.addEventListener("open", () => {
            console.log("BuzzerClient: opened")
            resolve(client)
        })
    })
}

const follow = (socket, followee) => {
    socket.send(["follow", followee].join(" "))
}

const getMessages = (socket, username) => {
    socket.send(["buzzfeed", username].join(" "))
}

const login = (socket, username, password) => (
    new Promise((resolve, reject) => {
        const response = (e) => {
            socket.removeEventListener("message", response)
            if (e.data === "OK")
                resolve()
            else
                reject(e.data)
        }
        socket.addEventListener("message", response)
        socket.send(["login", username, password].join(" "))
    })
)

const post = (socket, message) => (
    new Promise((resolve, reject) => {
        const response = (e) => {
            socket.removeEventListener("message", response)
            if (e.data.slice(0, 2) === "OK")
                resolve()
            else if (e.data.slice(0, 6) === "error ")
                reject(e.data)
            // Otherwise we are looking at an unrelated message.
        }
        socket.addEventListener("message", response)
        socket.send(["post", message].join(" "))
    })
)

const register = (socket, username, password) => (
    new Promise((resolve, reject) => {
        const response = (e) => {
            socket.removeEventListener("message", response)
            if (e.data === "OK")
                resolve()
            else
                reject(e.data)
        }
        socket.addEventListener("message", response)
        socket.send(["register", username, password].join(" "))
    })
)

const tagged = (socket, topic) => {
    socket.send(["topic", topic].join(' '))
}

const unfollow = (socket, followee) => {
    socket.send(["unfollow", followee].join(" "))
}


/// Parsing

const modes = {
    READY: Symbol(''),
    MENTION: Symbol('@'),
    TAG: Symbol('#'),
    WITHIN: Symbol('w')
}

const isWordChar = ch => /\w/.test(ch)

function linkify(msg, onClick) {
    const children = []
    let text = ""
    let mode = modes.READY

    msg += '\0'
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
                    children.push(
                        React.createElement("a", { href: "#mention", onClick }, text)
                    )
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
                    children.push(
                        React.createElement("a", { href: "#tag", onClick }, text)
                    )
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


/// React/UI Components

const LoginForm = (props) => (
    <form className="LoginForm" onSubmit={props.onSubmit}>
        <input name="username" autoComplete="username" placeholder="Username"
            defaultValue={props.username}
            disabled={props.disabled}
        />
        <input name="password" type="password" autoComplete="current-password"
            placeholder="Password" defaultValue={props.password}
            disabled={props.disabled}
        />
        <button type="submit" disabled={props.disabled}>{!props.disabled ? "Log In" : "..."}</button>
        <p className="center">
            New to Buzzer? <a href="#" onClick={props.onCancel}>Sign up!</a>
        </p>
    </form>
)

const RegistrationForm = (props) => (
    <form className="LoginForm" onSubmit={props.onSubmit}>
        <input name="username" autoComplete="username" placeholder="Username"
            defaultValue={props.username}
            disabled={props.disabled}
        />
        <input name="password" type="password" autoComplete="current-password"
            placeholder="Password" defaultValue={props.password}
            disabled={props.disabled}
        />
        <button type="submit" disabled={props.disabled}>{!props.disabled ? "Register" : "..."}</button>
        <p className="center">
            Already registered? <a href="#" onClick={props.onCancel}>Log in</a>
        </p>

    </form>
)

const endpoint = `ws://${window.location.host}/ws`
ReactDOM.render(
    <BuzzerWebView server={endpoint} />,
    document.getElementById("app")
)
