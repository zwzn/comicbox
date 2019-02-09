import autobind from 'autobind-decorator'
import * as s from 'css/login.scss'
import { login } from 'js/auth'
import Form from 'js/components/form'
import { Component, h } from 'preact'
import Btn from 'preact-material-components/Button'
import TextField from 'preact-material-components/TextField'
import { route } from 'preact-router'

export default class Login extends Component {

    public render() {
        return <div class={s.form}>
            <h1>Login</h1>
            <Form submit={this.login}>

                <TextField
                    class={s.username}
                    label='username'
                    name='user'
                />
                <TextField
                    class={s.password}
                    label='password'
                    type='password'
                    name='pass'
                />
                <Btn raised type='submit' class={s.login}>Login</Btn>
            </Form>
        </div >
    }

    @autobind
    private async login(data: { user: string, pass: string }) {

        const me = await login(data.user, data.pass)
        if (me === null) {
            alert('Invalid username or password')
            return
        }
        route('/')
    }
}
