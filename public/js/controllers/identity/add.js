define([
	'app',
	'factories/identity',
	'services/identities'
], function(app) {
	app.registerController('AddIdentityCtrl', ['$scope', '$state', 'Identity', function($scope, $state, Identity) {
		$scope.identity = new Identity();
		
		$scope.add = function () {
			$scope.identity.add()
			.success(function (data) {
				$state.transitionTo('identity.view', {
					identity_id: data._id
				});
			})
			.error(function (data) {

			})
		}
	}]);
});