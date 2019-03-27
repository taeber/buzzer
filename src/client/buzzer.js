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
        }

        this.getClient = this.getClient.bind(this)
        this.getMessages = this.getMessages.bind(this)
        this.search = this.search.bind(this)

        this.handleLogin = this.handleLogin.bind(this)
        this.handleLogout = this.handleLogout.bind(this)
        this.handleMessage = this.handleMessage.bind(this)
        this.handlePost = this.handlePost.bind(this)
        this.handleRegister = this.handleRegister.bind(this)
        this.handleSearch = this.handleSearch.bind(this)
        this.handleToggleForms = this.handleToggleForms.bind(this)
    }

    componentDidMount() {
        const banner = document.getElementById("banner")
        banner.addEventListener("click", () => {
            this.setState({ compressed: !this.state.compressed })
        })
    }

    render() {
        const {
            handleLogin, handleLogout, handlePost, handleRegister,
            handleSearch, handleToggleForms,
        } = this

        const {
            loggedIn, loginFormDisabled, username, messages, showRegistration,
            status, compressed,
        } = this.state

        if (showRegistration) {
            return (
                <RegistrationForm
                    disabled={loginFormDisabled}
                    onSubmit={handleRegister}
                    onCancel={handleToggleForms}
                />
            )
        }

        if (!loggedIn) {
            return (
                <LoginForm
                    disabled={loginFormDisabled}
                    onSubmit={handleLogin}
                    onCancel={handleToggleForms}
                />
            )
        }

        const messageList = (
            <ul key="messages" className="messages">
                {messages.map(msg => (
                    <li className="message" key={msg.id}>
                        <div className="poster">
                            @{msg.poster.username}
                            <span className="posted">
                                {moment(msg.posted).fromNow()}
                            </span>
                        </div>
                        <p className="text">{msg.text}</p>
                    </li>
                ))}
            </ul>
        )

        if (compressed)
            return messageList

        const hero = (
            <div key="hero" className="hero">
                <span className="big">@{username}</span>
                <a href="#" className="small" onClick={handleLogout}>Log out</a>
            </div>
        )

        const post = (
            <form key="post" className="PostForm" onSubmit={handlePost}>
                <input
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
        )

        const search = (
            <form key="search" className="SearchForm" onSubmit={handleSearch}>
                <input name="query" placeholder="#tag or @username" />
                <button type="submit" name="post">Search</button>
            </form>
        )

        return [hero, post, search, messageList]
    }

    async getClient() {
        let { client } = this.state
        if (!client) {
            client = await makeBuzzerClient(this.props.server, this.handleMessage)
            this.setState({ client })
        }
        return client
    }

    async getMessages() {
        this.setState({
            messages: [
                {
                    id: -2,
                    poster: { username: "therealbobross" },
                    text: "Happy Trees Happy Trees!",
                    posted: "2019-03-24T13:30:00-04:00",
                },
                {
                    id: -1,
                    poster: { username: "therealbobross" },
                    text: "This would be a good home for my little squirrels :)",
                    posted: "2019-03-24T13:30:01-04:00",
                }
            ]
        })
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
            }, this.getMessages)

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
        if (msg.slice(0, 5) !== "buzz ")
            return

        const buzz = JSON.parse(msg.slice(5))
        this.setState({
            messages: this.state.messages.concat([buzz])
        })
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
            }, this.getMessages)
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
    handleToggleForms(event) {
        event.preventDefault()
        this.setState({
            showRegistration: !this.state.showRegistration,
        })
    }

    search(query) {
        if (!query)
            return

        if (query[0] !== '#' || query[0] !== '@') {
            setTimeout(() => alert("Error! Query must start with # or @."), 1)
            return
        }
    }
}

function makeBuzzerClient(server, msgHandler) {
    return new Promise((resolve) => {
        const ws = new WebSocket(server)
        const client = {
            ws,
            Register: register.bind(null, ws),
            Login: login.bind(null, ws),
            Post: post.bind(null, ws),
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

const LoginForm = (props) => (
    <form className="LoginForm" onSubmit={props.onSubmit}>
        <input name="username" autoComplete="username" placeholder="Username" />
        <input name="password" type="password" autoComplete="current-password" placeholder="Password" />
        <button type="submit" disabled={props.disabled}>{!props.disabled ? "Log In" : "..."}</button>
        <p className="center">
            New to Buzzer? <a href="#" onClick={props.onCancel}>Sign up!</a>
        </p>
    </form>
)

const RegistrationForm = (props) => (
    <form className="LoginForm" onSubmit={props.onSubmit}>
        <input name="username" autoComplete="username" placeholder="Username" />
        <input name="password" type="password" autoComplete="current-password" placeholder="Password" />
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
