import React from 'react';
import MenuComponent from '../menu/menu';
import InvocationComponent from '../invocation/invocation';
import HomeComponent from '../home/home';
import capabilities from '../capabilities/capabilities'
import router, { Path } from '../router/router';
import authService, { AuthService, User } from '../auth/auth_service';

const denseModeKey = "VIEW_MODE";
const denseModeValue = "DENSE";

interface State {
  user: User;
  hash: string;
  path: string;
  search: URLSearchParams;
  denseMode: boolean;
}

export default class RootComponent extends React.Component {
  state: State = {
    user: null,
    hash: window.location.hash,
    path: window.location.pathname,
    search: new URLSearchParams(window.location.search),
    denseMode: window.localStorage.getItem(denseModeKey) == denseModeValue || false
  };

  componentWillMount() {
    capabilities.register("BuildBuddy Community Edition", [Path.invocationPath]);
    router.register(this.handlePathChange.bind(this));
    authService.userStream.addListener(AuthService.userEventName, (user: User) => {
      this.setState({ ...this.state, user })
    })
  }

  handlePathChange() {
    this.setState({
      hash: window.location.hash,
      path: window.location.pathname,
      search: new URLSearchParams(window.location.search)
    });
  }

  handleToggleDenseClicked() {
    let newDenseMode = !this.state.denseMode;
    this.setState({ ...this.state, denseMode: newDenseMode });
    window.localStorage.setItem(denseModeKey, newDenseMode ? denseModeValue : "");
  }

  render() {
    let invocationId = router.getInvocationId(this.state.path);
    return (
      <div className={this.state.denseMode ? "dense" : ""}>
        <MenuComponent user={this.state.user} denseModeEnabled={this.state.denseMode} handleDenseModeToggled={this.handleToggleDenseClicked.bind(this)} />
        {invocationId && <InvocationComponent invocationId={invocationId} hash={this.state.hash} search={this.state.search} denseMode={this.state.denseMode} />}
        {!invocationId && <HomeComponent />}
      </div>
    );
  }
}
