import autobind from 'autobind-decorator'
import * as s from 'css/layout.scss'
import { historyPop, historyPrevious } from 'js/history'
import { Component, h } from 'preact'
import 'preact-material-components/Button/style.css'
import Drawer from 'preact-material-components/Drawer'
import 'preact-material-components/Drawer/style.css'
import Icon from 'preact-material-components/Icon'
import List from 'preact-material-components/List'
import 'preact-material-components/List/style.css'
import TopAppBar from 'preact-material-components/TopAppBar'
import 'preact-material-components/TopAppBar/style.css'
import { Link, route } from 'preact-router'

interface Props {
    backLink: string
}

interface State {
    drawerOpened: boolean
}

export default class Layout extends Component<Props, State> {

    private searchInput: HTMLInputElement

    private get menu() {
        return [
            {
                name: 'Home',
                icon: 'home',
                href: '/',
            },
            {
                name: 'Lists',
                icon: 'view_list',
                href: '/current',
            },
            {
                name: 'Series',
                icon: 'collections_bookmark',
                href: '/series',
            },
            {
                name: 'Notifications',
                icon: 'notifications',
                href: '/notifications',
            },
            {
                name: 'Settings',
                icon: 'settings',
                href: '/settings',
            },
        ]
    }

    constructor() {
        super()
        this.state = {
            drawerOpened: false,
        }
    }

    public render() {
        let backButton = <TopAppBar.Icon />
        // <TopAppBar.Icon onClick={this.toggleDrawer} navigation={true}>menu</TopAppBar.Icon>

        if (location.hash !== '#/') {
            backButton = <TopAppBar.Icon onClick={this.btnBack} href='#' navigation={true}>
                arrow_back
            </TopAppBar.Icon>
        }

        return <div className={s.app}>
            <TopAppBar onNav={null}>
                <TopAppBar.Row>
                    <TopAppBar.Section align-start={true}>
                        {backButton}
                        <TopAppBar.Title>
                            <Link href='/'>ComicBox</Link>
                        </TopAppBar.Title>
                    </TopAppBar.Section>
                    <TopAppBar.Section align-end={true} class={s.search}>
                        <form onSubmit={this.search}>
                            <input id='search' type='text' ref={e => this.searchInput = e} />
                            <label for='search'>
                                <Icon>search</Icon>
                            </label>
                        </form>
                    </TopAppBar.Section>
                </TopAppBar.Row>
            </TopAppBar>

            {/* <Drawer modal={true} open={this.state.drawerOpened} onClose={this.drawerClosed}>
                <Drawer.DrawerHeader className='mdc-theme--primary-bg'>
                    Drawer Header
                </Drawer.DrawerHeader>
                <Drawer.DrawerContent>
                    <List>
                        {this.menu.map((item, i) =>
                            <Link key={i} href={item.href} onClick={this.toggleDrawer}>
                                <List.LinkItem
                                    className={location.pathname === item.href ? 'mdc-list-item--activated' : ''}
                                >
                                    <List.ItemGraphic>{item.icon}</List.ItemGraphic>
                                    {item.name}
                                </List.LinkItem>
                            </Link>,
                        )}
                    </List>
                </Drawer.DrawerContent>
            </Drawer> */}

            <main class={s.main}>
                {this.props.children}
            </main>

            <div class={s.bottomBar}>
                {this.menu.map((item, i) => <Link
                    key={i}
                    href={item.href}
                    class={s.link + ' ' + (location.hash === '#' + item.href ? s.active : '')}
                >
                    <Icon class={s.icon}>{item.icon}</Icon>
                    <div class={s.title}>{item.name}</div>
                </Link>)}
            </div>
        </div >
    }

    // @autobind
    // private toggleDrawer() {
    //     this.setState({
    //         drawerOpened: !this.state.drawerOpened,
    //     })
    // }

    // @autobind
    // private drawerClosed() {
    //     this.setState({
    //         drawerOpened: false,
    //     })
    // }

    @autobind
    private btnBack(e: Event) {
        e.preventDefault()
        if (historyPrevious() !== null) {
            historyPop()
            history.back()

            return
        }
        route(this.props.backLink)
    }

    @autobind
    private search(e: Event) {
        e.preventDefault()
        this.searchInput.blur()
        route(`/search?query=${this.searchInput.value}`)
    }
}
