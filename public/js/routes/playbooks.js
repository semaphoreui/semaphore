define([
	'app'
], function(app) {
	app.config(function($stateProvider) {
		$stateProvider
		.state('addPlaybook', {
			url: '/add',
			pageTitle: 'Add Playbook',
			templateUrl: "/view/playbook/add"
		})
	})
})