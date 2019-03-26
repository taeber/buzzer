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
            showRegistration: false,
            username: "",
            password: "",
            messages: [],
        }

        this.getMessages = this.getMessages.bind(this)
        this.search = this.search.bind(this)

        this.handleLogin = this.handleLogin.bind(this)
        this.handleLogout = this.handleLogout.bind(this)
        this.handlePost = this.handlePost.bind(this)
        this.handleRegister = this.handleRegister.bind(this)
        this.handleSearch = this.handleSearch.bind(this)
        this.handleToggleForms = this.handleToggleForms.bind(this)
    }

    render() {
        const {
            handleLogin, handleLogout, handlePost, handleRegister,
            handleSearch, handleToggleForms,
        } = this

        const {
            loggedIn, loginFormDisabled, username, messages, showRegistration,
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

        const hero = (
            <div key="hero" className="hero">
                <span className="big">@{username}</span>
                <a href="#" className="small" onClick={handleLogout}>Log out</a>
            </div>
        )

        const post = (
            <form key="post" className="PostForm" onSubmit={handlePost}>
                <input name="status" placeholder="What do you wanna say?" />
                <button type="submit" name="post">Post</button>
            </form>
        )

        const search = (
            <form key="search" className="SearchForm" onSubmit={handleSearch}>
                <input name="query" placeholder="#tag or @username" />
                <button type="submit" name="post">Search</button>
            </form>
        )

        const messageList = (
            <ul key="messages" className="messages">
                {messages.map(msg => (
                    <li className="message" key={msg.id}>
                        <div className="poster">
                            @{msg.poster}
                            <span className="posted">
                                {moment(msg.posted).fromNow()}
                            </span>
                        </div>
                        <p className="text">{msg.text}</p>
                    </li>
                ))}
            </ul>
        )

        return [hero, post, search, messageList]
    }

    async getMessages() {
        const response = await fetch("/").then(r => r.text())
        this.setState({
            messages: [
                {
                    id: 1,
                    poster: "therealbobross",
                    text: "Happy Trees Happy Trees!",
                    posted: "2019-03-24T13:30:00-04:00",
                },
                {
                    id: 2,
                    poster: "therealbobross",
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

        const creds = {
            username: document.getElementsByName("username")[0].value,
            password: document.getElementsByName("password")[0].value,
        }

        const response = await fetch("/").then(r => r.text())
        this.setState({
            username: creds.username,
            password: creds.password,
            loginFormDisabled: false,
            loggedIn: true,
            showRegistration: false,
        }, this.getMessages)
    }

    /**
     * @param {Event} event
     */
    handleLogout(event) {
        event.preventDefault()
        this.setState({
            loggedIn: false,
            password: "",
        })
    }

    /**
     * @param {Event} event
     */
    handlePost(event) {
        event.preventDefault()
    }

    /**
     * @param {Event} event
     */
    async handleRegister(event) {
        event.preventDefault()

        let { client } = this.state

        this.setState({ loginFormDisabled: true })

        const creds = {
            username: document.getElementsByName("username")[0].value,
            password: document.getElementsByName("password")[0].value,
        }

        if (!client) {
            client = await makeBuzzerClient(this.props.server)
            this.setState({ client })
        }

        try {
            await client.Register(creds.username, creds.password)
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

function makeBuzzerClient(server) {
    return new Promise((resolve) => {
        const ws = new WebSocket(server)
        const client = {
            ws,
            Register: register.bind(null, ws),
        }

        ws.onclose = () => console.log("BuzzerClient: closed")
        ws.onmessage = (e) => console.log("BuzzerClient: recv:", e.data)
        ws.onopen = () => resolve(client)
    })
}

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
            <a href="#" onClick={props.onCancel}>Cancel</a>
        </p>

    </form>
)

ReactDOM.render(
    <BuzzerWebView server="ws://localhost:8080/ws" />,
    document.getElementById("app")
)
