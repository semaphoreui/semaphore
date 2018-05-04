define(function () {
	app.registerController('ProjectEditCtrl', ['$scope', '$http', 'Project', '$state', 'SweetAlert', function ($scope, $http, Project, $state, SweetAlert) {
		$scope.projectName = Project.name;
		$scope.alert = Project.alert;
		$scope.alert_chat = Project.alert_chat;

		$scope.save = function (name, alert, alert_chat) {
			$http.put(Project.getURL(), {name: name, alert: alert, alert_chat: alert_chat}).then(function () {
				SweetAlert.swal('Saved', 'Project settings saved.', 'success');
			}).catch(function () {
				SweetAlert.swal('Error', 'Project settings were not saved', 'error');
			});
		}

		$scope.deleteProject = function () {
			SweetAlert.swal({
				title: 'Delete Project?',
				text: 'All data related to this project will be deleted.',
				icon: 'warning',
				buttons: {
					cancel: true,
					confirm: {
						text: 'Yes, DELETE',
						closeModal: false,
						className: 'bg-danger',
					},
				},
			}).then(function (value) {
				if (!value) {
					return;
				}

				$http.delete(Project.getURL())
					.then(function () {
						swal.stopLoading();
						swal.close();

						$state.go('dashboard');
					}).catch(function () {
					SweetAlert.swal('Error', 'Could not delete project!', 'error');
				});
			});
		}
	}]);
});