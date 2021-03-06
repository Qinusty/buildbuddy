import React from 'react';
import InvocationModel from './invocation_model'
import router from '../router/router';
import format from '../format/format';

interface Props {
  model: InvocationModel,
  invocationId: string
}
export default class InvocationOverviewComponent extends React.Component {
  props: Props;

  handleUserClicked() {
    router.navigateToUserHistory(this.props.model.getUser());
  }

  handleHostClicked() {
    router.navigateToHostHistory(this.props.model.getHost());
  }

  render() {
    return <div className="container">
      <div className="invocation">Invocation {this.props.invocationId}</div>
      <div className="titles">
        <div className="title">{format.sentenceCase(this.props.model.getUser())}'s {this.props.model.getCommand()} {this.props.model.getPattern()}</div>
        <div className="subtitle">{this.props.model.getStartDate()} at {this.props.model.getStartTime()}</div>
      </div>
      <div className="details">
        <div className="detail">
          {this.props.model.getStatusIcon()}
          {this.props.model.getStatus()}
        </div>
        <div className="detail" title={this.props.model.getDuractionSeconds()}>
          <img className="icon" src="/image/clock-regular.svg" />
          {this.props.model.getTiming()}
        </div>
        <div className="detail clickable" onClick={this.handleUserClicked.bind(this)}>
          <img className="icon" src="/image/user-regular.svg" />
          {this.props.model.getUser()}
        </div>
        <div className="detail clickable" onClick={this.handleHostClicked.bind(this)}>
          <img className="icon" src="/image/hard-drive-regular.svg" />
          {this.props.model.getHost()}
        </div>
        <div className="detail">
          <img className="icon" src="/image/tool-regular.svg" />
          {this.props.model.getTool()}
        </div>
        <div className="detail">
          <img className="icon" src="/image/grid-regular.svg" />
          {this.props.model.getPattern()}
        </div>
        <div className="detail" title={`${this.props.model.buildMetrics?.targetMetrics.targetsConfigured} configured / ${this.props.model.buildMetrics?.targetMetrics.targetsLoaded} loaded`}>
          <img className="icon" src="/image/target-regular.svg" />
          {this.props.model.targets.length} {this.props.model.targets.length == 1 ? "target" : "targets"}
        </div>
        <div title={`${this.props.model.buildMetrics?.actionSummary.actionsCreated} created`} className="detail">
          <img className="icon" src="/image/activity-regular.svg" />
          {this.props.model.buildMetrics?.actionSummary.actionsExecuted} actions
        </div>
        <div className="detail">
          <img className="icon" src="/image/box-regular.svg" />
          {this.props.model.buildMetrics?.packageMetrics.packagesLoaded} packages
        </div>
        <div className="detail">
          <img className="icon" src="/image/cpu-regular.svg" />
          {this.props.model.getCPU()}
        </div>
        <div className="detail">
          <img className="icon" src="/image/zap-regular.svg" />
          {this.props.model.getMode()}
        </div>
      </div>
    </div>
  }
}
