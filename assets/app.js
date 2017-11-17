$(function(){

	showVideos()

	function showVideos() {
		$('#error').hide()
		$('#results').empty()
		$.ajax({
			url: '/all-videos',
			success: render,
			error: function () {
				console.warn(arguments)
				$('#error').text("Oops, something went wrong - please refresh to try again").show()
			}
		})
	}

	function render(response) {
		// console.info(response)
		for (var i in response.items) {
			var item = response.items[i]
			console.info(item)

			$("#results").append(
				$("<a>", {
					'href': "/check?name=" + encodeURIComponent(item.name),
					'class': 'ui label',									
				}).text(item.name).click(processVideo)
			)			
		}
	}

	var framesContainer = $('#frames-container')
	var frames = $('#frames')
	var progress = $('.ui.progress').progress({
		total: 100,
		value: 0
	})
	var time = $('.time')
	var frameWidth = 150

	var frameData = {}

	function processVideo(e) {
		e.preventDefault()
		var $this = $(this)
		progress.show()
		frames.empty()
		$('.ui.dark.segment').slideDown()
		frameData = {}
		var framesWidth = 0
		
		var es = new EventSource($this.attr('href'))
		es.onmessage = function(e){
			
			var obj = JSON.parse(e.data)
			if (obj.thumbnail) {
				frameData[obj.frame] = obj
				frames.append(
					$("<img>", {
						src: 'data:image/jpg;base64,'+obj.thumbnail,
						'data-frame': obj.frame,
					}).click(selectFrame).click()
				)
				framesWidth += frameWidth
				frames.width(framesWidth)
				framesContainer.animate({
					scrollLeft: framesWidth - framesContainer.width(),
				}, 500)
			}	

			time.text(obj.seconds)
			progress.progress({
				total: obj.total_frames,
				value: obj.frame
			})

			if (obj.complete) {
				es.close()
				progress.slideUp()
				return
			}

		}

	}

	function selectFrame(e) {
		e.preventDefault()
		var $this = $(this)
		$('.ui.dark.segment').find('.ui.grid').fadeIn()
		var obj = frameData[$this.attr('data-frame')]
		console.info(obj)
		var facesEl = $('.faces').empty()
		var detailsEl = $('.details').empty().html('Frame: ' + obj.frame + ' at ' + obj.seconds)
		$('.frame-image').attr('src', $this.attr('src'))
		for (var i in obj.faces) {
			var face = obj.faces[i]
			facesEl.append(
				$('<li>').text(face.Name)
			)
		}
	}

})
